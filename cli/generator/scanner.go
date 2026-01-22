package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lite-lake/litecore-go/cli/analyzer"
)

// Scanner 扫描器
type Scanner struct {
	projectPath string
	moduleName  string
	analyzer    *analyzer.Analyzer
}

// NewScanner 创建扫描器
func NewScanner(projectPath, moduleName string) *Scanner {
	return &Scanner{
		projectPath: projectPath,
		moduleName:  moduleName,
		analyzer:    analyzer.NewAnalyzer(projectPath, moduleName),
	}
}

// Scan 扫描项目
func (s *Scanner) Scan() (*analyzer.ProjectInfo, error) {
	info, err := s.analyzer.Analyze()
	if err != nil {
		return nil, fmt.Errorf("analyze failed: %w", err)
	}

	s.scanComponents(info)

	return info, nil
}

type fileSet struct {
	files map[string]bool
	mu    sync.Mutex
}

func newFileSet() *fileSet {
	return &fileSet{
		files: make(map[string]bool),
	}
}

func (fs *fileSet) add(file string) bool {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.files[file] {
		return false
	}
	fs.files[file] = true
	return true
}

// scanComponents 扫描各层组件
func (s *Scanner) scanComponents(info *analyzer.ProjectInfo) {
	layerDirs := map[analyzer.Layer]string{
		analyzer.LayerEntity:     "entities",
		analyzer.LayerRepository: "repositories",
		analyzer.LayerService:    "services",
		analyzer.LayerController: "controllers",
		analyzer.LayerMiddleware: "middlewares",
	}

	for layer, dirName := range layerDirs {
		dirPath := filepath.Join(s.projectPath, "internal", dirName)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		files := s.findGoFiles(dirPath)
		for _, file := range files {
			s.analyzeComponentFile(file, layer, info)
		}
	}
}

// analyzeComponentFile 分析组件文件
func (s *Scanner) analyzeComponentFile(filename string, layer analyzer.Layer, info *analyzer.ProjectInfo) {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return
	}

	components, err := s.analyzer.Analyze()
	if err == nil {
		if layerComponents, exists := components.Layers[layer]; exists {
			for _, comp := range layerComponents {
				if comp.FileName == filename {
					s.extractFactoryFunc(comp, node)
				}
			}
		}
	}
}

// extractFactoryFunc 提取工厂函数
func (s *Scanner) extractFactoryFunc(comp *analyzer.ComponentInfo, node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == "New"+strings.TrimPrefix(comp.InterfaceName, "I") {
				comp.FactoryFunc = fn.Name.Name
			}
		}
		return true
	})
}

// findGoFiles 查找 Go 文件
func (s *Scanner) findGoFiles(dir string) []string {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil
	}

	return files
}

// getPackagePath 获取包路径
func (s *Scanner) getPackagePath(filename string) string {
	relPath := strings.TrimPrefix(filename, s.projectPath)
	relPath = strings.TrimPrefix(relPath, "/")
	relPath = strings.TrimPrefix(relPath, "\\")
	relPath = strings.TrimSuffix(relPath, filepath.Base(filename))

	parts := strings.Split(relPath, string(filepath.Separator))

	if len(parts) > 1 {
		return s.moduleName + "/" + strings.Join(parts[:len(parts)-1], "/")
	}

	return s.moduleName
}

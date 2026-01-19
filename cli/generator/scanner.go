package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"com.litelake.litecore/cli/analyzer"
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

	s.scanConfig()
	s.scanManagers()
	s.scanComponents(info)

	return info, nil
}

// scanConfig 扫描配置层
func (s *Scanner) scanConfig() {
	configDir := filepath.Join(s.projectPath, "configs")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return
	}

	configFiles := s.findGoFiles(configDir)
	for _, file := range configFiles {
		s.analyzeConfigFile(file)
	}
}

// analyzeConfigFile 分析配置文件
func (s *Scanner) analyzeConfigFile(filename string) {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == "NewConfigProvider" || strings.HasPrefix(fn.Name.Name, "NewConfig") {
				_ = &analyzer.ComponentInfo{
					InterfaceName: fn.Name.Name,
					InterfaceType: fn.Name.Name,
					PackagePath:   s.getPackagePath(filename),
					FileName:      filename,
					FactoryFunc:   fn.Name.Name,
					Layer:         analyzer.LayerConfig,
				}
			}
		}
		return true
	})
}

// scanManagers 扫描管理器层
func (s *Scanner) scanManagers() {
	managerDir := filepath.Join(s.projectPath, "manager")
	if _, err := os.Stat(managerDir); os.IsNotExist(err) {
		return
	}

	managerFiles := s.findGoFiles(managerDir)
	for _, file := range managerFiles {
		s.analyzeManagerFile(file)
	}
}

// analyzeManagerFile 分析管理器文件
func (s *Scanner) analyzeManagerFile(filename string) {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return
	}

	pkgName := ""
	ast.Inspect(node, func(n ast.Node) bool {
		if file, ok := n.(*ast.File); ok {
			pkgName = file.Name.Name
		}
		if fn, ok := n.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "BuildWithConfigProvider") || strings.HasPrefix(fn.Name.Name, "New") {
				_ = &analyzer.ComponentInfo{
					InterfaceName: fn.Name.Name,
					InterfaceType: pkgName + "." + fn.Name.Name,
					PackagePath:   s.getPackagePath(filename),
					FileName:      filename,
					FactoryFunc:   fn.Name.Name,
					Layer:         analyzer.LayerManager,
				}
			}
		}
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if strings.HasSuffix(typeSpec.Name.Name, "Manager") || strings.HasSuffix(typeSpec.Name.Name, "Mgr") {
				if _, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					_ = &analyzer.ComponentInfo{
						InterfaceName: typeSpec.Name.Name,
						InterfaceType: pkgName + "." + typeSpec.Name.Name,
						PackagePath:   s.getPackagePath(filename),
						FileName:      filename,
						Layer:         analyzer.LayerManager,
					}
				}
			}
		}
		return true
	})
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

package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lite-lake/litecore-go/cli/analyzer"
)

// Parser 解析器
type Parser struct {
	projectPath string
	info        *analyzer.ProjectInfo
}

// NewParser 创建解析器
func NewParser(projectPath string) *Parser {
	return &Parser{
		projectPath: projectPath,
		info: &analyzer.ProjectInfo{
			Layers: make(map[analyzer.Layer][]*analyzer.ComponentInfo),
		},
	}
}

// Parse 解析项目
func (p *Parser) Parse(moduleName string) (*analyzer.ProjectInfo, error) {
	p.info.ModuleName = moduleName

	if err := p.parseEntities(); err != nil {
		return nil, fmt.Errorf("parse entities failed: %w", err)
	}

	if err := p.parseInfras(); err != nil {
		return nil, fmt.Errorf("parse infras failed: %w", err)
	}

	if err := p.parseRepositories(); err != nil {
		return nil, fmt.Errorf("parse repositories failed: %w", err)
	}

	if err := p.parseServices(); err != nil {
		return nil, fmt.Errorf("parse services failed: %w", err)
	}

	if err := p.parseControllers(); err != nil {
		return nil, fmt.Errorf("parse controllers failed: %w", err)
	}

	if err := p.parseMiddlewares(); err != nil {
		return nil, fmt.Errorf("parse middlewares failed: %w", err)
	}

	if err := p.parseListeners(); err != nil {
		return nil, fmt.Errorf("parse listeners failed: %w", err)
	}

	if err := p.parseSchedulers(); err != nil {
		return nil, fmt.Errorf("parse schedulers failed: %w", err)
	}

	return p.info, nil
}

// parseEntities 解析实体
func (p *Parser) parseEntities() error {
	dir := filepath.Join(p.projectPath, "internal", "entities")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	files, err := p.findGoFiles(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := p.parseEntityFile(file); err != nil {
			return err
		}
	}

	return nil
}

// parseEntityFile 解析实体文件
func (p *Parser) parseEntityFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	_ = node.Name.Name
	packagePath := p.getPackagePath(filename)

	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if !ast.IsExported(typeSpec.Name.Name) {
				return true
			}

			if p.isEntityType(typeSpec.Type) {
				comp := &analyzer.ComponentInfo{
					InterfaceName: typeSpec.Name.Name,
					InterfaceType: typeSpec.Name.Name,
					PackagePath:   packagePath,
					FileName:      filename,
					Layer:         analyzer.LayerEntity,
				}
				p.info.Layers[analyzer.LayerEntity] = append(p.info.Layers[analyzer.LayerEntity], comp)
			}
		}
		return true
	})

	return nil
}

// isEntityType 判断是否为实体类型
func (p *Parser) isEntityType(expr ast.Expr) bool {
	if _, ok := expr.(*ast.StructType); ok {
		return true
	}
	return false
}

// parseRepositories 解析仓储
func (p *Parser) parseRepositories() error {
	dir := filepath.Join(p.projectPath, "internal", "repositories")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	files, err := p.findGoFiles(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := p.parseRepositoryFile(file); err != nil {
			return err
		}
	}

	return nil
}

// parseRepositoryFile 解析仓储文件
func (p *Parser) parseRepositoryFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgName := node.Name.Name
	packagePath := p.getPackagePath(filename)
	componentMap := make(map[string]*analyzer.ComponentInfo)

	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if strings.HasPrefix(typeSpec.Name.Name, "I") {
				comp := &analyzer.ComponentInfo{
					InterfaceName: typeSpec.Name.Name,
					InterfaceType: pkgName + "." + typeSpec.Name.Name,
					PackagePath:   packagePath,
					FileName:      filename,
					Layer:         analyzer.LayerRepository,
				}
				componentMap[typeSpec.Name.Name] = comp
			}
		}

		if fn, ok := n.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "New") {
				interfaceName := strings.TrimPrefix(fn.Name.Name, "New")
				if comp, exists := componentMap["I"+interfaceName]; exists && comp.FactoryFunc == "" {
					comp.FactoryFunc = fn.Name.Name
				}
			}
		}

		return true
	})

	for _, comp := range componentMap {
		p.info.Layers[analyzer.LayerRepository] = append(p.info.Layers[analyzer.LayerRepository], comp)
	}

	return nil
}

// parseServices 解析服务
func (p *Parser) parseServices() error {
	dir := filepath.Join(p.projectPath, "internal", "services")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	files, err := p.findGoFiles(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := p.parseServiceFile(file); err != nil {
			return err
		}
	}

	return nil
}

// parseServiceFile 解析服务文件
func (p *Parser) parseServiceFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgName := node.Name.Name
	packagePath := p.getPackagePath(filename)
	componentMap := make(map[string]*analyzer.ComponentInfo)

	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if strings.HasPrefix(typeSpec.Name.Name, "I") {
				comp := &analyzer.ComponentInfo{
					InterfaceName: typeSpec.Name.Name,
					InterfaceType: pkgName + "." + typeSpec.Name.Name,
					PackagePath:   packagePath,
					FileName:      filename,
					Layer:         analyzer.LayerService,
				}
				componentMap[typeSpec.Name.Name] = comp
			}
		}

		if fn, ok := n.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "New") {
				interfaceName := strings.TrimPrefix(fn.Name.Name, "New")
				if comp, exists := componentMap["I"+interfaceName]; exists && comp.FactoryFunc == "" {
					comp.FactoryFunc = fn.Name.Name
				}
			}
		}

		return true
	})

	for _, comp := range componentMap {
		p.info.Layers[analyzer.LayerService] = append(p.info.Layers[analyzer.LayerService], comp)
	}

	return nil
}

// parseControllers 解析控制器
func (p *Parser) parseControllers() error {
	dir := filepath.Join(p.projectPath, "internal", "controllers")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	files, err := p.findGoFiles(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := p.parseControllerFile(file); err != nil {
			return err
		}
	}

	return nil
}

// parseControllerFile 解析控制器文件
func (p *Parser) parseControllerFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgName := node.Name.Name
	packagePath := p.getPackagePath(filename)
	componentMap := make(map[string]*analyzer.ComponentInfo)

	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if strings.HasPrefix(typeSpec.Name.Name, "I") {
				comp := &analyzer.ComponentInfo{
					InterfaceName: typeSpec.Name.Name,
					InterfaceType: pkgName + "." + typeSpec.Name.Name,
					PackagePath:   packagePath,
					FileName:      filename,
					Layer:         analyzer.LayerController,
				}
				componentMap[typeSpec.Name.Name] = comp
			}
		}

		if fn, ok := n.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "New") {
				interfaceName := strings.TrimPrefix(fn.Name.Name, "New")
				if comp, exists := componentMap["I"+interfaceName]; exists && comp.FactoryFunc == "" {
					comp.FactoryFunc = fn.Name.Name
				}
			}
		}

		return true
	})

	for _, comp := range componentMap {
		p.info.Layers[analyzer.LayerController] = append(p.info.Layers[analyzer.LayerController], comp)
	}

	return nil
}

// parseMiddlewares 解析中间件
func (p *Parser) parseMiddlewares() error {
	dir := filepath.Join(p.projectPath, "internal", "middlewares")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	files, err := p.findGoFiles(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := p.parseMiddlewareFile(file); err != nil {
			return err
		}
	}

	return nil
}

// parseListeners 解析监听器
func (p *Parser) parseListeners() error {
	dir := filepath.Join(p.projectPath, "internal", "listeners")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	files, err := p.findGoFiles(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := p.parseListenerFile(file); err != nil {
			return err
		}
	}

	return nil
}

// parseMiddlewareFile 解析中间件文件
func (p *Parser) parseMiddlewareFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgName := node.Name.Name
	packagePath := p.getPackagePath(filename)
	componentMap := make(map[string]*analyzer.ComponentInfo)

	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if strings.HasPrefix(typeSpec.Name.Name, "I") {
				comp := &analyzer.ComponentInfo{
					InterfaceName: typeSpec.Name.Name,
					InterfaceType: pkgName + "." + typeSpec.Name.Name,
					PackagePath:   packagePath,
					FileName:      filename,
					Layer:         analyzer.LayerMiddleware,
				}
				componentMap[typeSpec.Name.Name] = comp
			}
		}

		if fn, ok := n.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "New") {
				interfaceName := strings.TrimPrefix(fn.Name.Name, "New")
				if comp, exists := componentMap["I"+interfaceName]; exists && comp.FactoryFunc == "" {
					comp.FactoryFunc = fn.Name.Name
				}
			}
		}

		return true
	})

	for _, comp := range componentMap {
		p.info.Layers[analyzer.LayerMiddleware] = append(p.info.Layers[analyzer.LayerMiddleware], comp)
	}

	return nil
}

// parseListenerFile 解析监听器文件
func (p *Parser) parseListenerFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgName := node.Name.Name
	packagePath := p.getPackagePath(filename)
	componentMap := make(map[string]*analyzer.ComponentInfo)

	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if strings.HasPrefix(typeSpec.Name.Name, "I") {
				comp := &analyzer.ComponentInfo{
					InterfaceName: typeSpec.Name.Name,
					InterfaceType: pkgName + "." + typeSpec.Name.Name,
					PackagePath:   packagePath,
					FileName:      filename,
					Layer:         analyzer.LayerListener,
				}
				componentMap[typeSpec.Name.Name] = comp
			}
		}

		if fn, ok := n.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "New") {
				interfaceName := strings.TrimPrefix(fn.Name.Name, "New")
				if comp, exists := componentMap["I"+interfaceName]; exists && comp.FactoryFunc == "" {
					comp.FactoryFunc = fn.Name.Name
				}
			}
		}

		return true
	})

	for _, comp := range componentMap {
		p.info.Layers[analyzer.LayerListener] = append(p.info.Layers[analyzer.LayerListener], comp)
	}

	return nil
}

// parseSchedulers 解析定时器
func (p *Parser) parseSchedulers() error {
	dir := filepath.Join(p.projectPath, "internal", "schedulers")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	files, err := p.findGoFiles(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := p.parseSchedulerFile(file); err != nil {
			return err
		}
	}

	return nil
}

// parseSchedulerFile 解析定时器文件
func (p *Parser) parseSchedulerFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgName := node.Name.Name
	packagePath := p.getPackagePath(filename)
	componentMap := make(map[string]*analyzer.ComponentInfo)

	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if strings.HasPrefix(typeSpec.Name.Name, "I") {
				comp := &analyzer.ComponentInfo{
					InterfaceName: typeSpec.Name.Name,
					InterfaceType: pkgName + "." + typeSpec.Name.Name,
					PackagePath:   packagePath,
					FileName:      filename,
					Layer:         analyzer.LayerScheduler,
				}
				componentMap[typeSpec.Name.Name] = comp
			}
		}

		if fn, ok := n.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "New") {
				interfaceName := strings.TrimPrefix(fn.Name.Name, "New")
				if comp, exists := componentMap["I"+interfaceName]; exists && comp.FactoryFunc == "" {
					comp.FactoryFunc = fn.Name.Name
				}
			}
		}

		return true
	})

	for _, comp := range componentMap {
		p.info.Layers[analyzer.LayerScheduler] = append(p.info.Layers[analyzer.LayerScheduler], comp)
	}

	return nil
}

// parseInfras 解析基础设施层
func (p *Parser) parseInfras() error {
	dir := filepath.Join(p.projectPath, "internal", "infras")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	files, err := p.findGoFiles(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := p.parseInfrasFile(file); err != nil {
			return err
		}
	}

	return nil
}

// parseInfrasFile 解析基础设施文件
func (p *Parser) parseInfrasFile(filename string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgName := node.Name.Name
	packagePath := p.getPackagePath(filename)

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if !strings.HasPrefix(fn.Name.Name, "New") || fn.Type.Results == nil {
				return true
			}

			if len(fn.Type.Results.List) == 0 {
				return true
			}

			typeName := strings.TrimPrefix(fn.Name.Name, "New")

			interfaceName := typeName
			interfaceType := pkgName + "." + typeName
			if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
				if selectorExpr, ok := fn.Type.Results.List[0].Type.(*ast.SelectorExpr); ok {
					if xIdent, ok := selectorExpr.X.(*ast.Ident); ok {
						interfaceType = xIdent.Name + "." + selectorExpr.Sel.Name
						interfaceName = selectorExpr.Sel.Name
					}
				} else if ident, ok := fn.Type.Results.List[0].Type.(*ast.Ident); ok {
					interfaceType = ident.Name
					interfaceName = ident.Name
				}
			}

			comp := &analyzer.ComponentInfo{
				InterfaceName: interfaceName,
				InterfaceType: interfaceType,
				PackagePath:   packagePath,
				FileName:      filename,
				FactoryFunc:   fn.Name.Name,
				Layer:         analyzer.LayerService,
			}

			p.info.Layers[analyzer.LayerService] = append(p.info.Layers[analyzer.LayerService], comp)
		}
		return true
	})

	return nil
}

// findGoFiles 查找 Go 文件
func (p *Parser) findGoFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// getPackagePath 获取包路径
func (p *Parser) getPackagePath(filename string) string {
	relPath := strings.TrimPrefix(filename, p.projectPath)
	relPath = strings.TrimPrefix(relPath, "/")
	relPath = strings.TrimPrefix(relPath, "\\")
	relPath = strings.TrimSuffix(relPath, filepath.Base(filename))

	parts := strings.Split(relPath, string(filepath.Separator))

	if len(parts) > 1 {
		return p.info.ModuleName + "/" + strings.Join(parts[:len(parts)-1], "/")
	}

	return p.info.ModuleName
}

// FindModuleName 查找模块名
func FindModuleName(projectPath string) (string, error) {
	goModPath := filepath.Join(projectPath, "go.mod")

	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`^module\s+(.+)`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid go.mod format")
	}

	return strings.TrimSpace(matches[1]), nil
}

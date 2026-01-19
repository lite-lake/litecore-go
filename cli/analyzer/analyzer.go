package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
)

// Layer 表示代码分层
type Layer string

const (
	LayerConfig     Layer = "config"
	LayerEntity     Layer = "entity"
	LayerManager    Layer = "manager"
	LayerRepository Layer = "repository"
	LayerService    Layer = "service"
	LayerController Layer = "controller"
	LayerMiddleware Layer = "middleware"
)

// ComponentInfo 组件信息
type ComponentInfo struct {
	InterfaceName string
	InterfaceType string
	PackagePath   string
	FileName      string
	FactoryFunc   string
	Layer         Layer
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	ModuleName  string
	PackagePath string
	ConfigPath  string
	Configs     []*ComponentInfo
	Managers    []*ComponentInfo
	Layers      map[Layer][]*ComponentInfo
}

// Analyzer 代码分析器
type Analyzer struct {
	projectPath string
	moduleName  string
	info        *ProjectInfo
}

// NewAnalyzer 创建分析器
func NewAnalyzer(projectPath, moduleName string) *Analyzer {
	return &Analyzer{
		projectPath: projectPath,
		moduleName:  moduleName,
		info: &ProjectInfo{
			ModuleName: moduleName,
			Layers:     make(map[Layer][]*ComponentInfo),
		},
	}
}

// Analyze 分析项目
func (a *Analyzer) Analyze() (*ProjectInfo, error) {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, a.projectPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse directory failed: %w", err)
	}

	for _, pkg := range pkgs {
		a.analyzePackage(pkg, fset)
	}

	return a.info, nil
}

// analyzePackage 分析包
func (a *Analyzer) analyzePackage(pkg *ast.Package, fset *token.FileSet) {
	for filename, file := range pkg.Files {
		a.analyzeFile(file, filename, fset)
	}
}

// analyzeFile 分析文件
func (a *Analyzer) analyzeFile(file *ast.File, filename string, fset *token.FileSet) {
	layer := a.detectLayer(filename, file.Name.Name)
	if layer == "" {
		return
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GenDecl:
			a.analyzeGenDecl(node, filename, layer, file.Name.Name, fset)
		case *ast.FuncDecl:
			a.analyzeFuncDecl(node, filename, layer, file.Name.Name)
		}
		return true
	})
}

// analyzeFuncDecl 分析函数声明（用于infras包中的工厂函数）
func (a *Analyzer) analyzeFuncDecl(fn *ast.FuncDecl, filename string, layer Layer, pkgName string) {
	if !strings.HasPrefix(fn.Name.Name, "New") || fn.Type.Results == nil {
		return
	}

	if len(fn.Type.Results.List) == 0 {
		return
	}

	typeName := strings.TrimPrefix(fn.Name.Name, "New")

	info := &ComponentInfo{
		InterfaceName: typeName,
		InterfaceType: pkgName + "." + typeName,
		PackagePath:   a.getPackagePath(filename),
		FileName:      filename,
		FactoryFunc:   fn.Name.Name,
		Layer:         layer,
	}

	a.info.Layers[layer] = append(a.info.Layers[layer], info)
}

// detectLayer 检测代码层
func (a *Analyzer) detectLayer(filename, packageName string) Layer {
	parts := strings.Split(filename, string("/"+"\\"))

	for _, part := range parts {
		if strings.Contains(part, "entities") {
			return LayerEntity
		}
		if strings.Contains(part, "repositories") {
			return LayerRepository
		}
		if strings.Contains(part, "services") {
			return LayerService
		}
		if strings.Contains(part, "controllers") {
			return LayerController
		}
		if strings.Contains(part, "middlewares") {
			return LayerMiddleware
		}
		if strings.Contains(part, "infras") {
			if strings.Contains(part, "configproviders") || strings.Contains(filename, "config_provider") {
				return LayerConfig
			}
			if strings.Contains(part, "managers") {
				return LayerManager
			}
			return LayerManager
		}
	}

	return ""
}

// analyzeGenDecl 分析通用声明
func (a *Analyzer) analyzeGenDecl(decl *ast.GenDecl, filename string, layer Layer, pkgName string, fset *token.FileSet) {
	for _, spec := range decl.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			a.analyzeTypeSpec(typeSpec, filename, layer, pkgName)
		}
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			a.analyzeValueSpec(valueSpec, filename, layer, pkgName)
		}
	}
}

// analyzeTypeSpec 分析类型声明
func (a *Analyzer) analyzeTypeSpec(typeSpec *ast.TypeSpec, filename string, layer Layer, pkgName string) {
	_, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return
	}

	interfaceName := typeSpec.Name.Name

	if strings.HasPrefix(interfaceName, "I") {
		info := &ComponentInfo{
			InterfaceName: interfaceName,
			InterfaceType: pkgName + "." + interfaceName,
			PackagePath:   a.getPackagePath(filename),
			FileName:      filename,
			Layer:         layer,
		}

		a.info.Layers[layer] = append(a.info.Layers[layer], info)
	}
}

// analyzeValueSpec 分析值声明
func (a *Analyzer) analyzeValueSpec(valueSpec *ast.ValueSpec, filename string, layer Layer, pkgName string) {
	for _, name := range valueSpec.Names {
		if name.Name == "default" || strings.HasSuffix(name.Name, "Impl") {
			continue
		}

		funcDecl := a.findFactoryFunc(filename, name.Name)
		if funcDecl == nil {
			continue
		}

		info := &ComponentInfo{
			InterfaceName: name.Name,
			InterfaceType: pkgName + "." + name.Name,
			PackagePath:   a.getPackagePath(filename),
			FileName:      filename,
			FactoryFunc:   "New" + name.Name,
			Layer:         layer,
		}

		a.info.Layers[layer] = append(a.info.Layers[layer], info)
	}
}

// findFactoryFunc 查找工厂函数
func (a *Analyzer) findFactoryFunc(filename, typeName string) *ast.FuncDecl {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil
	}

	var found *ast.FuncDecl

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == "New"+typeName {
				found = fn
				return false
			}
		}
		return true
	})

	return found
}

// getPackagePath 获取包路径
func (a *Analyzer) getPackagePath(filename string) string {
	relPath := strings.TrimPrefix(filename, a.projectPath)
	relPath = strings.TrimPrefix(relPath, "/")
	relPath = strings.TrimPrefix(relPath, "\\")
	relPath = strings.TrimSuffix(relPath, "/"+reflect.ValueOf("").String())
	relPath = strings.TrimSuffix(relPath, "\\"+reflect.ValueOf("").String())

	parts := strings.Split(relPath, string("/"+"\\"))

	if len(parts) > 1 {
		return a.moduleName + "/" + strings.Join(parts[:len(parts)-1], "/")
	}

	return a.moduleName
}

// IsLitecoreLayer 判断是否为 Litecore 标准层
func IsLitecoreLayer(layer Layer) bool {
	switch layer {
	case LayerConfig, LayerEntity, LayerManager,
		LayerRepository, LayerService, LayerController, LayerMiddleware:
		return true
	default:
		return false
	}
}

// GetBaseInterface 获取层对应的基础接口
func GetBaseInterface(layer Layer) string {
	switch layer {
	case LayerConfig:
		return "BaseConfigProvider"
	case LayerEntity:
		return "BaseEntity"
	case LayerManager:
		return "BaseManager"
	case LayerRepository:
		return "BaseRepository"
	case LayerService:
		return "BaseService"
	case LayerController:
		return "BaseController"
	case LayerMiddleware:
		return "BaseMiddleware"
	default:
		return ""
	}
}

// GetContainerName 获取容器名称
func GetContainerName(layer Layer) string {
	switch layer {
	case LayerConfig:
		return "ConfigContainer"
	case LayerEntity:
		return "EntityContainer"
	case LayerManager:
		return "ManagerContainer"
	case LayerRepository:
		return "RepositoryContainer"
	case LayerService:
		return "ServiceContainer"
	case LayerController:
		return "ControllerContainer"
	case LayerMiddleware:
		return "MiddlewareContainer"
	default:
		return ""
	}
}

// GetRegisterFunction 获取注册函数名
func GetRegisterFunction(layer Layer) string {
	switch layer {
	case LayerConfig:
		return "RegisterConfig"
	case LayerEntity:
		return "RegisterEntity"
	case LayerManager:
		return "RegisterManager"
	case LayerRepository:
		return "RegisterRepository"
	case LayerService:
		return "RegisterService"
	case LayerController:
		return "RegisterController"
	case LayerMiddleware:
		return "RegisterMiddleware"
	default:
		return "Register"
	}
}

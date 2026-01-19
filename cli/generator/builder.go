package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"com.litelake.litecore/cli/analyzer"
)

// Builder 代码生成器
type Builder struct {
	projectPath string
	outputDir   string
	packageName string
	moduleName  string
	configPath  string
}

// NewBuilder 创建生成器
func NewBuilder(projectPath, outputDir, packageName, moduleName, configPath string) *Builder {
	return &Builder{
		projectPath: projectPath,
		outputDir:   outputDir,
		packageName: packageName,
		moduleName:  moduleName,
		configPath:  configPath,
	}
}

// Generate 生成所有容器代码
func (b *Builder) Generate(info *analyzer.ProjectInfo) error {
	if err := b.generateConfigContainer(info); err != nil {
		return fmt.Errorf("generate config container failed: %w", err)
	}

	if err := b.generateEntityContainer(info); err != nil {
		return fmt.Errorf("generate entity container failed: %w", err)
	}

	if err := b.generateManagerContainer(info); err != nil {
		return fmt.Errorf("generate manager container failed: %w", err)
	}

	if err := b.generateRepositoryContainer(info); err != nil {
		return fmt.Errorf("generate repository container failed: %w", err)
	}

	if err := b.generateServiceContainer(info); err != nil {
		return fmt.Errorf("generate service container failed: %w", err)
	}

	if err := b.generateControllerContainer(info); err != nil {
		return fmt.Errorf("generate controller container failed: %w", err)
	}

	if err := b.generateMiddlewareContainer(info); err != nil {
		return fmt.Errorf("generate middleware container failed: %w", err)
	}

	if err := b.generateEngine(info); err != nil {
		return fmt.Errorf("generate engine failed: %w", err)
	}

	return nil
}

// generateConfigContainer 生成配置容器代码
func (b *Builder) generateConfigContainer(info *analyzer.ProjectInfo) error {
	components := b.convertComponents(info.Layers[analyzer.LayerConfig])

	data := &TemplateData{
		PackageName: b.packageName,
		ConfigPath:  b.configPath,
		Imports:     b.collectImports(info, analyzer.LayerConfig),
		Components:  components,
	}

	code, err := GenerateConfigContainer(data)
	if err != nil {
		return err
	}

	return b.writeFile("config_container.go", code)
}

// generateEntityContainer 生成实体容器代码
func (b *Builder) generateEntityContainer(info *analyzer.ProjectInfo) error {
	components := b.convertComponents(info.Layers[analyzer.LayerEntity])

	data := &TemplateData{
		PackageName: b.packageName,
		Imports:     b.collectImports(info, analyzer.LayerEntity),
		Components:  components,
	}

	code, err := GenerateEntityContainer(data)
	if err != nil {
		return err
	}

	return b.writeFile("entity_container.go", code)
}

// generateManagerContainer 生成管理器容器代码
func (b *Builder) generateManagerContainer(info *analyzer.ProjectInfo) error {
	components := b.convertComponents(info.Layers[analyzer.LayerManager])

	data := &TemplateData{
		PackageName: b.packageName,
		Imports:     b.collectImports(info, analyzer.LayerManager),
		Components:  components,
	}

	code, err := GenerateManagerContainer(data)
	if err != nil {
		return err
	}

	return b.writeFile("manager_container.go", code)
}

// generateRepositoryContainer 生成仓储容器代码
func (b *Builder) generateRepositoryContainer(info *analyzer.ProjectInfo) error {
	components := b.convertComponents(info.Layers[analyzer.LayerRepository])

	data := &TemplateData{
		PackageName: b.packageName,
		Imports:     b.collectImports(info, analyzer.LayerRepository),
		Components:  components,
	}

	code, err := GenerateRepositoryContainer(data)
	if err != nil {
		return err
	}

	return b.writeFile("repository_container.go", code)
}

// generateServiceContainer 生成服务容器代码
func (b *Builder) generateServiceContainer(info *analyzer.ProjectInfo) error {
	components := b.convertComponents(info.Layers[analyzer.LayerService])

	data := &TemplateData{
		PackageName: b.packageName,
		Imports:     b.collectImports(info, analyzer.LayerService),
		Components:  components,
	}

	code, err := GenerateServiceContainer(data)
	if err != nil {
		return err
	}

	return b.writeFile("service_container.go", code)
}

// generateControllerContainer 生成控制器容器代码
func (b *Builder) generateControllerContainer(info *analyzer.ProjectInfo) error {
	components := b.convertComponents(info.Layers[analyzer.LayerController])

	data := &TemplateData{
		PackageName: b.packageName,
		Imports:     b.collectImports(info, analyzer.LayerController),
		Components:  components,
	}

	code, err := GenerateControllerContainer(data)
	if err != nil {
		return err
	}

	return b.writeFile("controller_container.go", code)
}

// generateMiddlewareContainer 生成中间件容器代码
func (b *Builder) generateMiddlewareContainer(info *analyzer.ProjectInfo) error {
	components := b.convertComponents(info.Layers[analyzer.LayerMiddleware])

	data := &TemplateData{
		PackageName: b.packageName,
		Imports:     b.collectImports(info, analyzer.LayerMiddleware),
		Components:  components,
	}

	code, err := GenerateMiddlewareContainer(data)
	if err != nil {
		return err
	}

	return b.writeFile("middleware_container.go", code)
}

// generateEngine 生成引擎代码
func (b *Builder) generateEngine(info *analyzer.ProjectInfo) error {
	data := &TemplateData{
		PackageName: b.packageName,
	}

	code, err := GenerateEngine(data)
	if err != nil {
		return err
	}

	return b.writeFile("engine.go", code)
}

// convertComponents 转换组件数据
func (b *Builder) convertComponents(components []*analyzer.ComponentInfo) []ComponentTemplateData {
	var result []ComponentTemplateData

	for _, comp := range components {
		typeName := strings.TrimPrefix(comp.InterfaceName, "I")
		if typeName == "" {
			typeName = comp.InterfaceName
		}

		packageAlias := b.getPackageAlias(comp.PackagePath)
		interfaceType := comp.InterfaceType
		if !strings.Contains(interfaceType, ".") {
			interfaceType = packageAlias + "." + comp.InterfaceName
		}

		result = append(result, ComponentTemplateData{
			TypeName:      typeName,
			InterfaceName: comp.InterfaceName,
			InterfaceType: interfaceType,
			PackagePath:   comp.PackagePath,
			PackageAlias:  packageAlias,
			FactoryFunc:   comp.FactoryFunc,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].InterfaceName < result[j].InterfaceName
	})

	return result
}

// getPackageAlias 获取包别名
func (b *Builder) getPackageAlias(packagePath string) string {
	if packagePath == "" {
		return ""
	}

	parts := strings.Split(packagePath, "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}

// collectImports 收集导入
func (b *Builder) collectImports(info *analyzer.ProjectInfo, layer analyzer.Layer) map[string]string {
	importMap := make(map[string]string)

	components := info.Layers[layer]
	for _, comp := range components {
		if comp.PackagePath != "" && comp.PackagePath != b.moduleName {
			alias := b.getPackageAlias(comp.PackagePath)
			importMap[alias] = comp.PackagePath
		}

		if strings.Contains(comp.InterfaceType, ".") {
			parts := strings.Split(comp.InterfaceType, ".")
			if len(parts) > 1 {
				pkg := parts[0]
				if pkg != b.moduleName && pkg != "" && pkg != "common" && pkg != "config" {
					fullPkg := "com.litelake.litecore/manager/" + pkg
					if pkg == "telemetrymgr" {
						fullPkg = "com.litelake.litecore/manager/telemetrymgr"
					}
					if _, exists := importMap[pkg]; !exists {
						importMap[pkg] = fullPkg
					}
				}
			}
		}
	}

	return importMap
}

// writeFile 写入文件
func (b *Builder) writeFile(filename, content string) error {
	if err := os.MkdirAll(b.outputDir, 0755); err != nil {
		return fmt.Errorf("create output directory failed: %w", err)
	}

	filePath := filepath.Join(b.outputDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}

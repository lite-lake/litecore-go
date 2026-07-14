package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lite-lake/litecore-go/cli/analyzer"
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
	if err := b.generateEntityContainer(info); err != nil {
		return fmt.Errorf("生成实体容器失败: %w", err)
	}

	if err := b.generateRepositoryContainer(info); err != nil {
		return fmt.Errorf("生成仓储容器失败: %w", err)
	}

	if err := b.generateServiceContainer(info); err != nil {
		return fmt.Errorf("生成服务容器失败: %w", err)
	}

	if err := b.generateControllerContainer(info); err != nil {
		return fmt.Errorf("生成控制器容器失败: %w", err)
	}

	if err := b.generateMiddlewareContainer(info); err != nil {
		return fmt.Errorf("生成中间件容器失败: %w", err)
	}

	if err := b.generateListenerContainer(info); err != nil {
		return fmt.Errorf("生成监听器容器失败: %w", err)
	}

	if err := b.generateSchedulerContainer(info); err != nil {
		return fmt.Errorf("生成定时器容器失败: %w", err)
	}

	if err := b.generateEngine(info); err != nil {
		return fmt.Errorf("生成引擎失败: %w", err)
	}

	return nil
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

// generateListenerContainer 生成监听器容器代码
func (b *Builder) generateListenerContainer(info *analyzer.ProjectInfo) error {
	components := b.convertComponents(info.Layers[analyzer.LayerListener])

	data := &TemplateData{
		PackageName: b.packageName,
		Imports:     b.collectImports(info, analyzer.LayerListener),
		Components:  components,
	}

	code, err := GenerateListenerContainer(data)
	if err != nil {
		return err
	}

	return b.writeFile("listener_container.go", code)
}

// generateSchedulerContainer 生成定时器容器代码
func (b *Builder) generateSchedulerContainer(info *analyzer.ProjectInfo) error {
	components := b.convertComponents(info.Layers[analyzer.LayerScheduler])

	data := &TemplateData{
		PackageName: b.packageName,
		Imports:     b.collectImports(info, analyzer.LayerScheduler),
		Components:  components,
	}

	code, err := GenerateSchedulerContainer(data)
	if err != nil {
		return err
	}

	return b.writeFile("scheduler_container.go", code)
}

// generateEngine 生成引擎代码
func (b *Builder) generateEngine(info *analyzer.ProjectInfo) error {
	configPath := b.configPath
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(b.projectPath, configPath)
	}

	absConfigPath, err := filepath.Abs(configPath)
	if err == nil {
		relConfigPath, err := filepath.Rel(b.projectPath, absConfigPath)
		if err == nil {
			configPath = filepath.ToSlash(relConfigPath)
		}
	}

	data := &TemplateData{
		PackageName: b.packageName,
		ConfigPath:  configPath,
	}

	code, err := GenerateEngine(data)
	if err != nil {
		return err
	}

	return b.writeFile("engine.go", code)
}

// convertComponents 转换组件数据
func (b *Builder) convertComponents(components []*analyzer.ComponentInfo) []ComponentTemplateData {
	seen := make(map[string]bool)
	var result []ComponentTemplateData

	for _, comp := range components {
		if comp.FactoryFunc == "" && comp.Layer != analyzer.LayerEntity {
			continue
		}

		key := comp.InterfaceName + ":" + comp.FileName
		if seen[key] {
			continue
		}
		seen[key] = true

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
			Layer:         string(comp.Layer),
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

// ImportEntry 导入条目
type ImportEntry struct {
	Alias string
	Path  string
	Group int
}

// collectImports 收集并排序导入
func (b *Builder) collectImports(info *analyzer.ProjectInfo, layer analyzer.Layer) []ImportEntry {
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
				if pkg != b.moduleName && pkg != "" && pkg != "common" && pkg != "configmgr" {
					fullPkg := "github.com/lite-lake/litecore-go/manager/" + pkg
					if pkg == "telemetrymgr" {
						fullPkg = "github.com/lite-lake/litecore-go/component/manager/telemetrymgr"
					}
					if _, exists := importMap[pkg]; !exists {
						importMap[pkg] = fullPkg
					}
				}
			}
		}
	}

	// 转换为 slice 并排序
	var entries []ImportEntry
	for alias, path := range importMap {
		group := b.getImportGroup(path)
		entries = append(entries, ImportEntry{Alias: alias, Path: path, Group: group})
	}

	// 分组排序
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Group != entries[j].Group {
			return entries[i].Group < entries[j].Group
		}
		return entries[i].Path < entries[j].Path
	})

	return entries
}

// getImportGroup 获取导入分组（用于排序）
// 0: 标准库, 1: 第三方库, 2: litecore, 3: 本地模块
func (b *Builder) getImportGroup(path string) int {
	// 标准库（不包含点）
	if !strings.Contains(path, ".") {
		return 0
	}
	// litecore
	if strings.HasPrefix(path, "github.com/lite-lake/litecore-go") {
		return 2
	}
	// 本地模块
	if strings.HasPrefix(path, b.moduleName) {
		return 3
	}
	// 第三方库
	return 1
}

// writeFile 写入文件
func (b *Builder) writeFile(filename, content string) error {
	if err := os.MkdirAll(b.outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	filePath := filepath.Join(b.outputDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

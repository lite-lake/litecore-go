package loggermgr

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	// 默认值
	DefaultMaxSize    = 100 // MB
	DefaultMaxAge     = 30  // days
	DefaultMaxBackups = 10
	MaxSafeSize       = 10000 // 最大安全值 10GB
	MaxSafeAge        = 3650  // 最大安全值 10年
	MaxSafeBackups    = 1000  // 最大安全备份数
)

// DefaultConfig 返回默认配置（使用 none 驱动）
func DefaultConfig() *LoggerConfig {
	return &LoggerConfig{
		Driver:    "none",
		ZapConfig: DefaultZapConfig(),
	}
}

// LoggerConfig 日志管理配置
type LoggerConfig struct {
	Driver    string     `yaml:"driver"`     // 驱动类型: zap, none
	ZapConfig *ZapConfig `yaml:"zap_config"` // Zap 驱动配置
}

// ZapConfig Zap 日志配置
type ZapConfig struct {
	TelemetryEnabled bool            `yaml:"telemetry_enabled"` // 是否启用观测日志
	TelemetryConfig  *LogLevelConfig `yaml:"telemetry_config"`  // 观测日志配置
	ConsoleEnabled   bool            `yaml:"console_enabled"`   // 是否启用控制台日志
	ConsoleConfig    *LogLevelConfig `yaml:"console_config"`    // 控制台日志配置
	FileEnabled      bool            `yaml:"file_enabled"`      // 是否启用文件日志
	FileConfig       *FileLogConfig  `yaml:"file_config"`       // 文件日志配置
}

// LogLevelConfig 日志级别配置
type LogLevelConfig struct {
	Level string `yaml:"level"` // 日志级别: debug, info, warn, error, fatal
}

// FileLogConfig 文件日志配置
type FileLogConfig struct {
	Level    string          `yaml:"level"`    // 日志级别
	Path     string          `yaml:"path"`     // 日志文件路径
	Rotation *RotationConfig `yaml:"rotation"` // 日志轮转配置
}

// RotationConfig 日志轮转配置
type RotationConfig struct {
	MaxSize    int  `yaml:"max_size"`    // 单个日志文件最大大小（MB），如 100MB
	MaxAge     int  `yaml:"max_age"`     // 日志文件保留天数，如 30d
	MaxBackups int  `yaml:"max_backups"` // 保留的旧日志文件最大数量
	Compress   bool `yaml:"compress"`    // 是否压缩旧日志文件
}

// DefaultZapConfig 返回默认的 Zap 配置
func DefaultZapConfig() *ZapConfig {
	return &ZapConfig{
		TelemetryEnabled: false,
		TelemetryConfig: &LogLevelConfig{
			Level: "info",
		},
		ConsoleEnabled: true,
		ConsoleConfig: &LogLevelConfig{
			Level: "info",
		},
		FileEnabled: false,
		FileConfig: &FileLogConfig{
			Level: "info",
			Rotation: &RotationConfig{
				MaxSize:    DefaultMaxSize,
				MaxAge:     DefaultMaxAge,
				MaxBackups: DefaultMaxBackups,
				Compress:   true,
			},
		},
	}
}

// Validate 验证配置
func (c *LoggerConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("driver is required")
	}

	// 标准化驱动名称
	c.Driver = strings.ToLower(strings.TrimSpace(c.Driver))

	switch c.Driver {
	case "zap", "none":
		// 有效驱动
	default:
		return fmt.Errorf("unsupported driver: %s (must be zap or none)", c.Driver)
	}

	// Zap 驱动需要 Zap 配置
	if c.Driver == "zap" && c.ZapConfig == nil {
		return fmt.Errorf("zap_config is required when driver is zap")
	}

	// 如果有 Zap 配置，验证它
	if c.ZapConfig != nil {
		if err := c.ZapConfig.Validate(); err != nil {
			return fmt.Errorf("invalid zap_config: %w", err)
		}
	}

	return nil
}

// Validate 验证 Zap 配置
func (c *ZapConfig) Validate() error {
	// 至少需要启用一种日志输出
	if !c.TelemetryEnabled && !c.ConsoleEnabled && !c.FileEnabled {
		return fmt.Errorf("at least one logger output must be enabled (telemetry, console, or file)")
	}

	// 验证观测日志配置
	if c.TelemetryEnabled && c.TelemetryConfig != nil {
		if !IsValidLogLevel(c.TelemetryConfig.Level) {
			return fmt.Errorf("invalid telemetry log level: %s", c.TelemetryConfig.Level)
		}
	}

	// 验证控制台日志配置
	if c.ConsoleEnabled && c.ConsoleConfig != nil {
		if !IsValidLogLevel(c.ConsoleConfig.Level) {
			return fmt.Errorf("invalid console log level: %s", c.ConsoleConfig.Level)
		}
	}

	// 验证文件日志配置
	if c.FileEnabled && c.FileConfig != nil {
		if !IsValidLogLevel(c.FileConfig.Level) {
			return fmt.Errorf("invalid file log level: %s", c.FileConfig.Level)
		}
		if c.FileConfig.Path == "" {
			return fmt.Errorf("file log path is required when file logging is enabled")
		}
	}

	return nil
}

// ParseLoggerConfigFromMap 从 ConfigMap 解析日志配置
func ParseLoggerConfigFromMap(cfg map[string]any) (*LoggerConfig, error) {
	config := &LoggerConfig{
		Driver:    "none", // 默认使用 none 驱动
		ZapConfig: DefaultZapConfig(),
	}

	if cfg == nil {
		return config, nil
	}

	// 解析 driver
	if driver, ok := cfg["driver"].(string); ok {
		config.Driver = strings.ToLower(strings.TrimSpace(driver))
	}

	// 解析 zap_config
	if zapConfigMap, ok := cfg["zap_config"].(map[string]any); ok {
		zapConfig, err := parseZapConfig(zapConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse zap_config: %w", err)
		}
		config.ZapConfig = zapConfig
	}

	return config, nil
}

// parseZapConfig 解析 Zap 配置
func parseZapConfig(cfg map[string]any) (*ZapConfig, error) {
	zapConfig := DefaultZapConfig()

	if cfg == nil {
		return zapConfig, nil
	}

	// 解析 telemetry_enabled
	if telemetryEnabled, ok := cfg["telemetry_enabled"].(bool); ok {
		zapConfig.TelemetryEnabled = telemetryEnabled
	}

	// 解析 telemetry_config
	if telemetryConfigMap, ok := cfg["telemetry_config"].(map[string]any); ok {
		telemetryConfig, err := parseLogLevelConfig(telemetryConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse telemetry_config: %w", err)
		}
		zapConfig.TelemetryConfig = telemetryConfig
	}

	// 解析 console_enabled
	if consoleEnabled, ok := cfg["console_enabled"].(bool); ok {
		zapConfig.ConsoleEnabled = consoleEnabled
	}

	// 解析 console_config
	if consoleConfigMap, ok := cfg["console_config"].(map[string]any); ok {
		consoleConfig, err := parseLogLevelConfig(consoleConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse console_config: %w", err)
		}
		zapConfig.ConsoleConfig = consoleConfig
	}

	// 解析 file_enabled
	if fileEnabled, ok := cfg["file_enabled"].(bool); ok {
		zapConfig.FileEnabled = fileEnabled
	}

	// 解析 file_config
	if fileConfigMap, ok := cfg["file_config"].(map[string]any); ok {
		fileConfig, err := parseFileLogConfig(fileConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file_config: %w", err)
		}
		zapConfig.FileConfig = fileConfig
	}

	return zapConfig, nil
}

// parseLogLevelConfig 解析日志级别配置
func parseLogLevelConfig(cfg map[string]any) (*LogLevelConfig, error) {
	config := &LogLevelConfig{Level: "info"}

	if level, ok := cfg["level"].(string); ok {
		config.Level = level
	}

	return config, nil
}

// parseFileLogConfig 解析文件日志配置
func parseFileLogConfig(cfg map[string]any) (*FileLogConfig, error) {
	config := &FileLogConfig{
		Level: "info",
		Rotation: &RotationConfig{
			MaxSize:    DefaultMaxSize,
			MaxAge:     DefaultMaxAge,
			MaxBackups: DefaultMaxBackups,
			Compress:   true,
		},
	}

	if level, ok := cfg["level"].(string); ok {
		config.Level = level
	}

	if path, ok := cfg["path"].(string); ok {
		config.Path = path
	}

	if rotationMap, ok := cfg["rotation"].(map[string]any); ok {
		rotation, err := parseRotationConfig(rotationMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rotation configmgr: %w", err)
		}
		config.Rotation = rotation
	}

	return config, nil
}

// parseRotationConfig 解析日志轮转配置
func parseRotationConfig(cfg map[string]any) (*RotationConfig, error) {
	config := &RotationConfig{
		MaxSize:    DefaultMaxSize,
		MaxAge:     DefaultMaxAge,
		MaxBackups: DefaultMaxBackups,
		Compress:   true,
	}

	// 解析 max_size
	if v, ok := cfg["max_size"]; ok {
		size, err := parseSizeValue(v)
		if err != nil {
			return nil, fmt.Errorf("invalid max_size: %w", err)
		}
		config.MaxSize = size
	}

	// 解析 max_age
	if v, ok := cfg["max_age"]; ok {
		age, err := parseAgeValue(v)
		if err != nil {
			return nil, fmt.Errorf("invalid max_age: %w", err)
		}
		config.MaxAge = age
	}

	// 解析 max_backups
	if v, ok := cfg["max_backups"]; ok {
		backups, err := parseBackupsValue(v)
		if err != nil {
			return nil, fmt.Errorf("invalid max_backups: %w", err)
		}
		config.MaxBackups = backups
	}

	// 解析 compress
	if compress, ok := cfg["compress"].(bool); ok {
		config.Compress = compress
	}

	return config, nil
}

// parseSizeValue 解析大小值，支持 int、int64、float64 和字符串
func parseSizeValue(v any) (int, error) {
	switch val := v.(type) {
	case int:
		return clampSize(val), nil
	case int64:
		// 检查是否超出 int 范围
		if val > math.MaxInt32 || val < math.MinInt32 {
			return DefaultMaxSize, fmt.Errorf("int64 value %d out of safe range", val)
		}
		return clampSize(int(val)), nil
	case float64:
		// 检查是否为整数且在安全范围内
		if val != float64(int64(val)) {
			return DefaultMaxSize, fmt.Errorf("float64 value %v is not an integer", val)
		}
		if val > math.MaxInt32 || val < 0 {
			return DefaultMaxSize, fmt.Errorf("float64 value %v out of safe range", val)
		}
		return clampSize(int(val)), nil
	case string:
		return parseSizeString(val)
	default:
		return DefaultMaxSize, fmt.Errorf("unsupported type %T for size", v)
	}
}

// parseSizeString 解析类似 "100MB" 或 "100" 的字符串
func parseSizeString(s string) (int, error) {
	s = strings.TrimSpace(strings.ToUpper(s))

	// 如果是纯数字，直接解析
	if num, err := strconv.Atoi(s); err == nil {
		return clampSize(num), nil
	}

	// 解析带单位的字符串
	if len(s) < 2 {
		return DefaultMaxSize, fmt.Errorf("size string too short: %s", s)
	}

	// 提取数字部分
	numStr := s[:len(s)-2]
	unit := s[len(s)-2:]

	num, err := strconv.Atoi(numStr)
	if err != nil {
		// 尝试整个字符串作为数字
		if num, err := strconv.Atoi(s); err == nil {
			return clampSize(num), nil
		}
		return DefaultMaxSize, fmt.Errorf("invalid size format: %s", s)
	}

	// 根据单位转换
	switch unit {
	case "MB":
		return clampSize(num), nil
	case "GB":
		return clampSize(num * 1024), nil
	case "KB":
		return clampSize(num / 1024), nil
	default:
		// 尝试不带单位的数字
		if num, err := strconv.Atoi(s); err == nil {
			return clampSize(num), nil
		}
		return DefaultMaxSize, fmt.Errorf("unsupported size unit: %s", unit)
	}
}

// parseAgeValue 解析时间值，支持 int、int64、float64 和字符串
func parseAgeValue(v any) (int, error) {
	switch val := v.(type) {
	case int:
		return clampAge(val), nil
	case int64:
		if val > math.MaxInt32 || val < math.MinInt32 {
			return DefaultMaxAge, fmt.Errorf("int64 value %d out of safe range", val)
		}
		return clampAge(int(val)), nil
	case float64:
		if val != float64(int64(val)) {
			return DefaultMaxAge, fmt.Errorf("float64 value %v is not an integer", val)
		}
		if val > math.MaxInt32 || val < 0 {
			return DefaultMaxAge, fmt.Errorf("float64 value %v out of safe range", val)
		}
		return clampAge(int(val)), nil
	case string:
		return parseAgeString(val)
	default:
		return DefaultMaxAge, fmt.Errorf("unsupported type %T for age", v)
	}
}

// parseAgeString 解析类似 "30d" 的字符串
func parseAgeString(s string) (int, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	// 如果是纯数字，直接解析
	if num, err := strconv.Atoi(s); err == nil {
		return clampAge(num), nil
	}

	if len(s) < 2 {
		return DefaultMaxAge, fmt.Errorf("age string too short: %s", s)
	}

	// 提取数字部分
	numStr := s[:len(s)-1]
	unit := s[len(s)-1:]

	num, err := strconv.Atoi(numStr)
	if err != nil {
		// 尝试整个字符串作为数字
		if num, err := strconv.Atoi(s); err == nil {
			return clampAge(num), nil
		}
		return DefaultMaxAge, fmt.Errorf("invalid age format: %s", s)
	}

	// 根据单位转换
	switch unit {
	case "d":
		return clampAge(num), nil
	case "h":
		return clampAge(num / 24), nil
	default:
		// 尝试不带单位的数字
		if num, err := strconv.Atoi(s); err == nil {
			return clampAge(num), nil
		}
		return DefaultMaxAge, fmt.Errorf("unsupported age unit: %s (use 'd' for days or 'h' for hours)", unit)
	}
}

// parseBackupsValue 解析备份数值
func parseBackupsValue(v any) (int, error) {
	switch val := v.(type) {
	case int:
		return clampBackups(val), nil
	case int64:
		if val > math.MaxInt32 || val < math.MinInt32 {
			return DefaultMaxBackups, fmt.Errorf("int64 value %d out of safe range", val)
		}
		return clampBackups(int(val)), nil
	case float64:
		if val != float64(int64(val)) {
			return DefaultMaxBackups, fmt.Errorf("float64 value %v is not an integer", val)
		}
		if val > math.MaxInt32 || val < 0 {
			return DefaultMaxBackups, fmt.Errorf("float64 value %v out of safe range", val)
		}
		return clampBackups(int(val)), nil
	case string:
		num, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil {
			return DefaultMaxBackups, fmt.Errorf("invalid backups value: %s", val)
		}
		return clampBackups(num), nil
	default:
		return DefaultMaxBackups, fmt.Errorf("unsupported type %T for backups", v)
	}
}

// clampSize 限制大小在安全范围内
func clampSize(size int) int {
	if size <= 0 {
		return DefaultMaxSize
	}
	if size > MaxSafeSize {
		return MaxSafeSize
	}
	return size
}

// clampAge 限制时间在安全范围内
func clampAge(age int) int {
	if age <= 0 {
		return DefaultMaxAge
	}
	if age > MaxSafeAge {
		return MaxSafeAge
	}
	return age
}

// clampBackups 限制备份数在安全范围内
func clampBackups(backups int) int {
	if backups < 0 {
		return DefaultMaxBackups
	}
	if backups > MaxSafeBackups {
		return MaxSafeBackups
	}
	return backups
}

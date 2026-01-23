package loggermgr

type Config struct {
	Driver    string           `yaml:"driver"`     // 驱动类型: zap, default, none
	ZapConfig *DriverZapConfig `yaml:"zap_config"` // Zap 驱动配置
}

type DriverZapConfig struct {
	TelemetryEnabled bool            `yaml:"telemetry_enabled"` // 是否启用观测日志
	TelemetryConfig  *LogLevelConfig `yaml:"telemetry_config"`  // 观测日志配置
	ConsoleEnabled   bool            `yaml:"console_enabled"`   // 是否启用控制台日志
	ConsoleConfig    *LogLevelConfig `yaml:"console_config"`    // 控制台日志配置
	FileEnabled      bool            `yaml:"file_enabled"`      // 是否启用文件日志
	FileConfig       *FileLogConfig  `yaml:"file_config"`       // 文件日志配置
}

type LogLevelConfig struct {
	Level string `yaml:"level"` // 日志级别: debug, info, warn, error, fatal
}

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

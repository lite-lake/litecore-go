package databasemgr

import (
	"fmt"
	"time"
)

const (
	// 默认连接池配置
	DefaultMaxOpenConns    = 10
	DefaultMaxIdleConns    = 5
	DefaultConnMaxLifetime = 30 * time.Second
	DefaultConnMaxIdleTime = 5 * time.Minute
)

// DatabaseConfig 数据库管理配置
type DatabaseConfig struct {
	Driver              string               `yaml:"driver"`               // 驱动类型: mysql, postgresql, sqlite, none
	SQLiteConfig        *SQLiteConfig        `yaml:"sqlite_config"`        // SQLite 配置
	PostgreSQLConfig    *PostgreSQLConfig    `yaml:"postgresql_config"`    // PostgreSQL 配置
	MySQLConfig         *MySQLConfig         `yaml:"mysql_config"`         // MySQL 配置
	ObservabilityConfig *ObservabilityConfig `yaml:"observability_config"` // 可观测性配置
	AutoMigrate         bool                 `yaml:"auto_migrate"`         // 是否自动迁移数据库表结构（默认 false）
}

// PoolConfig 数据库连接池配置（所有驱动通用）
type PoolConfig struct {
	MaxOpenConns    int           `yaml:"max_open_conns"`     // 最大打开连接数，0 表示无限制
	MaxIdleConns    int           `yaml:"max_idle_conns"`     // 最大空闲连接数
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`  // 连接最大存活时间
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"` // 连接最大空闲时间
}

// SQLiteConfig SQLite 配置
type SQLiteConfig struct {
	DSN        string      `yaml:"dsn"`         // SQLite DSN，如: file:./data.db?cache=shared&mode=rwc
	PoolConfig *PoolConfig `yaml:"pool_config"` // 连接池配置（可选）
}

// PostgreSQLConfig PostgreSQL 配置
type PostgreSQLConfig struct {
	DSN string `yaml:"dsn"` // PostgreSQL DSN，如: host=localhost port=5432 user=postgres password=
	// password dbname=lite_demo sslmode=disable
	PoolConfig *PoolConfig `yaml:"pool_config"` // 连接池配置（可选）
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	DSN string `yaml:"dsn"` // MySQL DSN，如: root:password@tcp(localhost:3306)/lite_demo?
	// charset=utf8mb4&parseTime=True&loc=Local
	PoolConfig *PoolConfig `yaml:"pool_config"` // 连接池配置（可选）
}

// ObservabilityConfig 可观测性配置
type ObservabilityConfig struct {
	// SlowQueryThreshold 慢查询阈值，0 表示不记录慢查询
	SlowQueryThreshold time.Duration `yaml:"slow_query_threshold"`

	// LogSQL 是否记录完整的 SQL 语句（生产环境建议关闭）
	LogSQL bool `yaml:"log_sql"`

	// SampleRate 采样率（0.0-1.0），1.0 表示全部记录
	SampleRate float64 `yaml:"sample_rate"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Driver:      "none",
		AutoMigrate: false,
		ObservabilityConfig: &ObservabilityConfig{
			SlowQueryThreshold: 1 * time.Second,
			LogSQL:             false,
			SampleRate:         1.0,
		},
	}
}

// Validate 验证配置
func (c *DatabaseConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("driver is required")
	}

	// 验证驱动类型
	if !isValidDriver(c.Driver) {
		return fmt.Errorf("invalid driver: %s, must be one of: mysql, postgresql, sqlite, none", c.Driver)
	}

	// 验证驱动对应的配置
	switch c.Driver {
	case "sqlite":
		if c.SQLiteConfig == nil {
			return fmt.Errorf("sqlite_config is required when driver is sqlite")
		}
		if err := c.SQLiteConfig.Validate(); err != nil {
			return fmt.Errorf("invalid sqlite_config: %w", err)
		}
	case "postgresql":
		if c.PostgreSQLConfig == nil {
			return fmt.Errorf("postgresql_config is required when driver is postgresql")
		}
		if err := c.PostgreSQLConfig.Validate(); err != nil {
			return fmt.Errorf("invalid postgresql_config: %w", err)
		}
	case "mysql":
		if c.MySQLConfig == nil {
			return fmt.Errorf("mysql_config is required when driver is mysql")
		}
		if err := c.MySQLConfig.Validate(); err != nil {
			return fmt.Errorf("invalid mysql_config: %w", err)
		}
	case "none":
		// none 驱动不需要配置
	}

	return nil
}

// Validate 验证 SQLite 配置
func (c *SQLiteConfig) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("sqlite DSN is required")
	}

	// 验证连接池配置
	if c.PoolConfig != nil {
		if err := c.PoolConfig.Validate(); err != nil {
			return fmt.Errorf("invalid pool_config: %w", err)
		}
	}

	return nil
}

// Validate 验证 PostgreSQL 配置
func (c *PostgreSQLConfig) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("postgresql DSN is required")
	}

	// 验证连接池配置
	if c.PoolConfig != nil {
		if err := c.PoolConfig.Validate(); err != nil {
			return fmt.Errorf("invalid pool_config: %w", err)
		}
	}

	return nil
}

// Validate 验证 MySQL 配置
func (c *MySQLConfig) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("mysql DSN is required")
	}

	// 验证连接池配置
	if c.PoolConfig != nil {
		if err := c.PoolConfig.Validate(); err != nil {
			return fmt.Errorf("invalid pool_config: %w", err)
		}
	}

	return nil
}

// Validate 验证连接池配置
func (c *PoolConfig) Validate() error {
	if c.MaxOpenConns < 0 {
		return fmt.Errorf("max_open_conns must be >= 0")
	}

	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns must be >= 0")
	}

	if c.MaxIdleConns > c.MaxOpenConns && c.MaxOpenConns > 0 {
		return fmt.Errorf("max_idle_conns must be <= max_open_conns")
	}

	if c.ConnMaxLifetime < 0 {
		return fmt.Errorf("conn_max_lifetime must be >= 0")
	}

	if c.ConnMaxIdleTime < 0 {
		return fmt.Errorf("conn_max_idle_time must be >= 0")
	}

	return nil
}

// isValidDriver 检查驱动是否有效
func isValidDriver(driver string) bool {
	switch driver {
	case "mysql", "postgresql", "sqlite", "none":
		return true
	default:
		return false
	}
}

// ParseDatabaseConfigFromMap 从 map 解析数据库配置
func ParseDatabaseConfigFromMap(cfg map[string]any) (*DatabaseConfig, error) {
	databaseConfig := DefaultConfig()

	if cfg == nil {
		return databaseConfig, nil
	}

	// 解析 driver
	if driver, ok := cfg["driver"].(string); ok {
		databaseConfig.Driver = driver
	}

	// 解析 sqlite_config
	if sqliteConfigMap, ok := cfg["sqlite_config"].(map[string]any); ok {
		sqliteConfig, err := parseSQLiteConfig(sqliteConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sqlite_config: %w", err)
		}
		databaseConfig.SQLiteConfig = sqliteConfig
	}

	// 解析 postgresql_config
	if postgresqlConfigMap, ok := cfg["postgresql_config"].(map[string]any); ok {
		postgresqlConfig, err := parsePostgreSQLConfig(postgresqlConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse postgresql_config: %w", err)
		}
		databaseConfig.PostgreSQLConfig = postgresqlConfig
	}

	// 解析 mysql_config
	if mysqlConfigMap, ok := cfg["mysql_config"].(map[string]any); ok {
		mysqlConfig, err := parseMySQLConfig(mysqlConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse mysql_config: %w", err)
		}
		databaseConfig.MySQLConfig = mysqlConfig
	}

	// 解析 observability_config
	if obsConfigMap, ok := cfg["observability_config"].(map[string]any); ok {
		obsConfig, err := parseObservabilityConfig(obsConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse observability_config: %w", err)
		}
		databaseConfig.ObservabilityConfig = obsConfig
	}

	// 解析 auto_migrate
	if autoMigrate, ok := cfg["auto_migrate"].(bool); ok {
		databaseConfig.AutoMigrate = autoMigrate
	}

	return databaseConfig, nil
}

// parseSQLiteConfig 解析 SQLite 配置
func parseSQLiteConfig(cfg map[string]any) (*SQLiteConfig, error) {
	config := &SQLiteConfig{
		PoolConfig: &PoolConfig{
			MaxOpenConns:    1, // SQLite 通常设置为 1
			MaxIdleConns:    1,
			ConnMaxLifetime: DefaultConnMaxLifetime,
			ConnMaxIdleTime: DefaultConnMaxIdleTime,
		},
	}

	if dsn, ok := cfg["dsn"].(string); ok {
		config.DSN = dsn
	}

	if poolConfigMap, ok := cfg["pool_config"].(map[string]any); ok {
		poolConfig, err := parsePoolConfig(poolConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pool_config: %w", err)
		}
		config.PoolConfig = poolConfig
	}

	return config, nil
}

// parsePostgreSQLConfig 解析 PostgreSQL 配置
func parsePostgreSQLConfig(cfg map[string]any) (*PostgreSQLConfig, error) {
	config := &PostgreSQLConfig{
		PoolConfig: &PoolConfig{
			MaxOpenConns:    DefaultMaxOpenConns,
			MaxIdleConns:    DefaultMaxIdleConns,
			ConnMaxLifetime: DefaultConnMaxLifetime,
			ConnMaxIdleTime: DefaultConnMaxIdleTime,
		},
	}

	if dsn, ok := cfg["dsn"].(string); ok {
		config.DSN = dsn
	}

	if poolConfigMap, ok := cfg["pool_config"].(map[string]any); ok {
		poolConfig, err := parsePoolConfig(poolConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pool_config: %w", err)
		}
		config.PoolConfig = poolConfig
	}

	return config, nil
}

// parseMySQLConfig 解析 MySQL 配置
func parseMySQLConfig(cfg map[string]any) (*MySQLConfig, error) {
	config := &MySQLConfig{
		PoolConfig: &PoolConfig{
			MaxOpenConns:    DefaultMaxOpenConns,
			MaxIdleConns:    DefaultMaxIdleConns,
			ConnMaxLifetime: DefaultConnMaxLifetime,
			ConnMaxIdleTime: DefaultConnMaxIdleTime,
		},
	}

	if dsn, ok := cfg["dsn"].(string); ok {
		config.DSN = dsn
	}

	if poolConfigMap, ok := cfg["pool_config"].(map[string]any); ok {
		poolConfig, err := parsePoolConfig(poolConfigMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse pool_config: %w", err)
		}
		config.PoolConfig = poolConfig
	}

	return config, nil
}

// parsePoolConfig 解析连接池配置
func parsePoolConfig(cfg map[string]any) (*PoolConfig, error) {
	config := &PoolConfig{
		MaxOpenConns:    DefaultMaxOpenConns,
		MaxIdleConns:    DefaultMaxIdleConns,
		ConnMaxLifetime: DefaultConnMaxLifetime,
		ConnMaxIdleTime: DefaultConnMaxIdleTime,
	}

	// 解析 max_open_conns
	if v, ok := cfg["max_open_conns"]; ok {
		if num, ok := v.(int); ok {
			config.MaxOpenConns = num
		} else if num, ok := v.(float64); ok {
			config.MaxOpenConns = int(num)
		}
	}

	// 解析 max_idle_conns
	if v, ok := cfg["max_idle_conns"]; ok {
		if num, ok := v.(int); ok {
			config.MaxIdleConns = num
		} else if num, ok := v.(float64); ok {
			config.MaxIdleConns = int(num)
		}
	}

	// 解析 conn_max_lifetime
	if v, ok := cfg["conn_max_lifetime"]; ok {
		if duration, ok := v.(int); ok {
			config.ConnMaxLifetime = time.Duration(duration) * time.Second
		} else if duration, ok := v.(float64); ok {
			config.ConnMaxLifetime = time.Duration(duration) * time.Second
		} else if durationStr, ok := v.(string); ok {
			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				return nil, fmt.Errorf("invalid conn_max_lifetime format: %s", durationStr)
			}
			config.ConnMaxLifetime = duration
		}
	}

	// 解析 conn_max_idle_time
	if v, ok := cfg["conn_max_idle_time"]; ok {
		if duration, ok := v.(int); ok {
			config.ConnMaxIdleTime = time.Duration(duration) * time.Second
		} else if duration, ok := v.(float64); ok {
			config.ConnMaxIdleTime = time.Duration(duration) * time.Second
		} else if durationStr, ok := v.(string); ok {
			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				return nil, fmt.Errorf("invalid conn_max_idle_time format: %s", durationStr)
			}
			config.ConnMaxIdleTime = duration
		}
	}

	return config, nil
}

// parseObservabilityConfig 解析可观测性配置
func parseObservabilityConfig(cfg map[string]any) (*ObservabilityConfig, error) {
	config := &ObservabilityConfig{
		SlowQueryThreshold: 1 * time.Second, // 默认 1 秒
		LogSQL:             false,           // 默认不记录 SQL
		SampleRate:         1.0,             // 默认全采样
	}

	// 解析 slow_query_threshold
	if v, ok := cfg["slow_query_threshold"]; ok {
		if duration, ok := v.(int); ok {
			config.SlowQueryThreshold = time.Duration(duration) * time.Second
		} else if duration, ok := v.(float64); ok {
			config.SlowQueryThreshold = time.Duration(duration) * time.Second
		} else if durationStr, ok := v.(string); ok {
			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				return nil, fmt.Errorf("invalid slow_query_threshold format: %s", durationStr)
			}
			config.SlowQueryThreshold = duration
		}
	}

	// 解析 log_sql
	if logSQL, ok := cfg["log_sql"].(bool); ok {
		config.LogSQL = logSQL
	}

	// 解析 sample_rate
	if sampleRate, ok := cfg["sample_rate"]; ok {
		if rate, ok := sampleRate.(float64); ok {
			if rate < 0 || rate > 1 {
				return nil, fmt.Errorf("sample_rate must be between 0 and 1, got: %f", rate)
			}
			config.SampleRate = rate
		}
	}

	return config, nil
}

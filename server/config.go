package server

import (
	"strconv"
	"time"
)

// StartupLogConfig 启动日志配置
type StartupLogConfig struct {
	Enabled bool `yaml:"enabled"` // 是否启用启动日志
	Async   bool `yaml:"async"`   // 是否异步输出（默认 true）
	Buffer  int  `yaml:"buffer"`  // 缓冲区大小（默认 100）
}

// DefaultStartupLogConfig 返回默认的启动日志配置
func DefaultStartupLogConfig() *StartupLogConfig {
	return &StartupLogConfig{
		Enabled: true,
		Async:   true,
		Buffer:  100,
	}
}

// serverConfig 服务器配置
type serverConfig struct {
	Host            string            // 监听地址，默认 0.0.0.0
	Port            int               // 监听端口，默认 8080
	Mode            string            // 运行模式：debug/release/test，默认 release
	ReadTimeout     time.Duration     // 读取超时，默认 10s
	WriteTimeout    time.Duration     // 写入超时，默认 10s
	IdleTimeout     time.Duration     // 空闲超时，默认 60s
	ShutdownTimeout time.Duration     // 关闭超时，默认 30s
	StartupLog      *StartupLogConfig // 启动日志配置
}

// defaultServerConfig 返回默认的服务器配置
func defaultServerConfig() *serverConfig {
	return &serverConfig{
		Host:            "0.0.0.0",
		Port:            8080,
		Mode:            "release",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		StartupLog:      DefaultStartupLogConfig(),
	}
}

// Address 返回服务器监听地址
func (c *serverConfig) Address() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

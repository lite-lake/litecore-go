package server

import (
	"fmt"
	"path/filepath"
	"strings"

	"com.litelake.litecore/common"
	"com.litelake.litecore/config"
)

// EngineOption 配置选项函数类型
type EngineOption func(*Engine)

// WithConfigFile 直接传入配置文件路径
// 支持格式：.json, .yaml, .yml
func WithConfigFile(path string) EngineOption {
	return func(e *Engine) {
		// 根据文件扩展名选择驱动
		ext := strings.ToLower(filepath.Ext(path))
		var driver string
		switch ext {
		case ".json":
			driver = "json"
		case ".yaml", ".yml":
			driver = "yaml"
		default:
			e.initErrors = append(e.initErrors, fmt.Errorf("unsupported config file format: %s", ext))
			return
		}

		provider, err := config.NewConfigProvider(driver, path)
		if err != nil {
			e.initErrors = append(e.initErrors, fmt.Errorf("failed to create config provider: %w", err))
			return
		}

		if err := e.containers.config.Register(provider); err != nil {
			e.initErrors = append(e.initErrors, fmt.Errorf("failed to register config: %w", err))
		}
	}
}

// WithConfig 直接传入配置对象
func WithConfig(cfg common.BaseConfigProvider) EngineOption {
	return func(e *Engine) {
		if err := e.containers.config.Register(cfg); err != nil {
			e.initErrors = append(e.initErrors, fmt.Errorf("failed to register config: %w", err))
		}
	}
}

// WithServerConfig 设置服务器配置
// 传入 nil 则使用默认配置
func WithServerConfig(cfg *ServerConfig) EngineOption {
	return func(e *Engine) {
		if cfg == nil {
			e.serverConfig = DefaultServerConfig()
		} else {
			e.serverConfig = cfg
		}
	}
}

// RegisterManagers 批量注册管理器
func RegisterManagers(managers ...interface{}) EngineOption {
	return func(e *Engine) {
		for _, mgr := range managers {
			if baseMgr, ok := mgr.(common.BaseManager); ok {
				if err := e.containers.manager.Register(baseMgr); err != nil {
					e.initErrors = append(e.initErrors, fmt.Errorf("failed to register manager: %w", err))
				}
			} else {
				e.initErrors = append(e.initErrors, fmt.Errorf("manager %T does not implement BaseManager interface", mgr))
			}
		}
	}
}

// RegisterEntities 批量注册实体
func RegisterEntities(entities ...interface{}) EngineOption {
	return func(e *Engine) {
		for _, ent := range entities {
			if baseEnt, ok := ent.(common.BaseEntity); ok {
				if err := e.containers.entity.Register(baseEnt); err != nil {
					e.initErrors = append(e.initErrors, fmt.Errorf("failed to register entity: %w", err))
				}
			} else {
				e.initErrors = append(e.initErrors, fmt.Errorf("entity %T does not implement BaseEntity interface", ent))
			}
		}
	}
}

// RegisterRepositories 批量注册仓储
func RegisterRepositories(repos ...interface{}) EngineOption {
	return func(e *Engine) {
		for _, repo := range repos {
			if baseRepo, ok := repo.(common.BaseRepository); ok {
				if err := e.containers.repository.Register(baseRepo); err != nil {
					e.initErrors = append(e.initErrors, fmt.Errorf("failed to register repository: %w", err))
				}
			} else {
				e.initErrors = append(e.initErrors, fmt.Errorf("repository %T does not implement BaseRepository interface", repo))
			}
		}
	}
}

// RegisterServices 批量注册服务
func RegisterServices(services ...interface{}) EngineOption {
	return func(e *Engine) {
		for _, svc := range services {
			if baseSvc, ok := svc.(common.BaseService); ok {
				if err := e.containers.service.Register(baseSvc); err != nil {
					e.initErrors = append(e.initErrors, fmt.Errorf("failed to register service: %w", err))
				}
			} else {
				e.initErrors = append(e.initErrors, fmt.Errorf("service %T does not implement BaseService interface", svc))
			}
		}
	}
}

// RegisterControllers 批量注册控制器
func RegisterControllers(controllers ...interface{}) EngineOption {
	return func(e *Engine) {
		for _, ctrl := range controllers {
			if baseCtrl, ok := ctrl.(common.BaseController); ok {
				if err := e.containers.controller.Register(baseCtrl); err != nil {
					e.initErrors = append(e.initErrors, fmt.Errorf("failed to register controller: %w", err))
				}
			} else {
				e.initErrors = append(e.initErrors, fmt.Errorf("controller %T does not implement BaseController interface", ctrl))
			}
		}
	}
}

// RegisterMiddlewares 批量注册中间件
func RegisterMiddlewares(middlewares ...interface{}) EngineOption {
	return func(e *Engine) {
		for _, mw := range middlewares {
			if baseMw, ok := mw.(common.BaseMiddleware); ok {
				if err := e.containers.middleware.Register(baseMw); err != nil {
					e.initErrors = append(e.initErrors, fmt.Errorf("failed to register middleware: %w", err))
				}
			} else {
				e.initErrors = append(e.initErrors, fmt.Errorf("middleware %T does not implement BaseMiddleware interface", mw))
			}
		}
	}
}

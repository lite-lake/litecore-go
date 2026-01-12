package server

import (
	"fmt"
	"path/filepath"
	"reflect"
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

		if err := e.containers.config.RegisterByType(reflect.TypeOf((*common.BaseConfigProvider)(nil)).Elem(), provider); err != nil {
			e.initErrors = append(e.initErrors, fmt.Errorf("failed to register config: %w", err))
		}
	}
}

// WithConfig 直接传入配置对象
func WithConfig(cfg common.BaseConfigProvider) EngineOption {
	return func(e *Engine) {
		if err := e.containers.config.RegisterByType(reflect.TypeOf((*common.BaseConfigProvider)(nil)).Elem(), cfg); err != nil {
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

// RegisterPair 注册对（用于泛型注册辅助）
type RegisterPair struct {
	ifaceType reflect.Type
	impl      interface{}
}

// Register 泛型注册辅助函数
// 用于在 EngineOption 中注册接口类型到实现
func Register[T any](impl interface{}) *RegisterPair {
	ifaceType := reflect.TypeOf((*T)(nil)).Elem()
	return &RegisterPair{
		ifaceType: ifaceType,
		impl:      impl,
	}
}

// RegisterManagers 批量注册管理器
func RegisterManagers(pairs ...*RegisterPair) EngineOption {
	return func(e *Engine) {
		for _, pair := range pairs {
			if pair == nil {
				continue
			}
			if baseMgr, ok := pair.impl.(common.BaseManager); ok {
				if err := e.containers.manager.RegisterByType(pair.ifaceType, baseMgr); err != nil {
					e.initErrors = append(e.initErrors, fmt.Errorf("failed to register manager: %w", err))
				}
			} else {
				e.initErrors = append(e.initErrors, fmt.Errorf("manager %T does not implement BaseManager interface", pair.impl))
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
func RegisterRepositories(pairs ...*RegisterPair) EngineOption {
	return func(e *Engine) {
		for _, pair := range pairs {
			if pair == nil {
				continue
			}
			if baseRepo, ok := pair.impl.(common.BaseRepository); ok {
				if err := e.containers.repository.RegisterByType(pair.ifaceType, baseRepo); err != nil {
					e.initErrors = append(e.initErrors, fmt.Errorf("failed to register repository: %w", err))
				}
			} else {
				e.initErrors = append(e.initErrors, fmt.Errorf("repository %T does not implement BaseRepository interface", pair.impl))
			}
		}
	}
}

// RegisterServices 批量注册服务
func RegisterServices(pairs ...*RegisterPair) EngineOption {
	return func(e *Engine) {
		for _, pair := range pairs {
			if pair == nil {
				continue
			}
			if baseSvc, ok := pair.impl.(common.BaseService); ok {
				if err := e.containers.service.RegisterByType(pair.ifaceType, baseSvc); err != nil {
					e.initErrors = append(e.initErrors, fmt.Errorf("failed to register service: %w", err))
				}
			} else {
				e.initErrors = append(e.initErrors, fmt.Errorf("service %T does not implement BaseService interface", pair.impl))
			}
		}
	}
}

// RegisterControllers 批量注册控制器
func RegisterControllers(pairs ...*RegisterPair) EngineOption {
	return func(e *Engine) {
		for _, pair := range pairs {
			if pair == nil {
				continue
			}
			if baseCtrl, ok := pair.impl.(common.BaseController); ok {
				if err := e.containers.controller.RegisterByType(pair.ifaceType, baseCtrl); err != nil {
					e.initErrors = append(e.initErrors, fmt.Errorf("failed to register controller: %w", err))
				}
			} else {
				e.initErrors = append(e.initErrors, fmt.Errorf("controller %T does not implement BaseController interface", pair.impl))
			}
		}
	}
}

// RegisterMiddlewares 批量注册中间件
func RegisterMiddlewares(pairs ...*RegisterPair) EngineOption {
	return func(e *Engine) {
		for _, pair := range pairs {
			if pair == nil {
				continue
			}
			if baseMw, ok := pair.impl.(common.BaseMiddleware); ok {
				if err := e.containers.middleware.RegisterByType(pair.ifaceType, baseMw); err != nil {
					e.initErrors = append(e.initErrors, fmt.Errorf("failed to register middleware: %w", err))
				}
			} else {
				e.initErrors = append(e.initErrors, fmt.Errorf("middleware %T does not implement BaseMiddleware interface", pair.impl))
			}
		}
	}
}

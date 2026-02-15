package litehtmltemplatesvc

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lite-lake/litecore-go/common"
	"github.com/lite-lake/litecore-go/manager/loggermgr"
)

type ILiteHTMLTemplateService interface {
	common.IBaseService
	Render(ctx *gin.Context, name string, data interface{})
	RenderWithCode(ctx *gin.Context, code int, name string, data interface{})
	SetGinEngine(engine *gin.Engine)
	ReloadTemplates() error
	AddFunc(name string, fn interface{})
	AddFuncMap(funcMap template.FuncMap)
	GetTemplateNames() []string
	HasTemplate(name string) bool
}

type I18nService interface {
	T(lang, key string) string
	TWithData(lang, key string, data map[string]interface{}) string
}

type liteHTMLTemplateServiceImpl struct {
	cfg       *Config
	ginEngine *gin.Engine
	LoggerMgr loggermgr.ILoggerManager `inject:""`
	I18nSvc   I18nService              `inject:""`
	funcMap   template.FuncMap
	templates *template.Template
	loaded    bool
}

func NewLiteHTMLTemplateService() ILiteHTMLTemplateService {
	return NewLiteHTMLTemplateServiceWithConfig(nil)
}

func NewLiteHTMLTemplateServiceWithConfig(config *Config) ILiteHTMLTemplateService {
	cfg := config
	if cfg == nil {
		cfg = &Config{}
	}

	s := &liteHTMLTemplateServiceImpl{
		cfg:     cfg,
		funcMap: make(template.FuncMap),
	}

	s.initBuiltinFuncs()

	if customFuncs := cfg.getFuncMap(); customFuncs != nil {
		for name, fn := range customFuncs {
			s.funcMap[name] = fn
		}
	}

	return s
}

func (s *liteHTMLTemplateServiceImpl) initBuiltinFuncs() {
	s.funcMap["lower"] = strings.ToLower
	s.funcMap["upper"] = strings.ToUpper
	s.funcMap["trim"] = strings.TrimSpace
	s.funcMap["safe"] = func(str string) template.HTML {
		return template.HTML(str)
	}
	s.funcMap["formatDate"] = func(t time.Time, layout string) string {
		return t.Format(layout)
	}
	s.funcMap["dict"] = func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("dict requires even number of arguments")
		}
		result := make(map[string]interface{})
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict key must be string")
			}
			result[key] = values[i+1]
		}
		return result, nil
	}
	s.funcMap["json"] = func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	s.funcMap["default"] = func(defaultValue, value interface{}) interface{} {
		if value == nil || value == "" {
			return defaultValue
		}
		return value
	}
}

func (s *liteHTMLTemplateServiceImpl) ServiceName() string {
	return "HTMLTemplateService"
}

func (s *liteHTMLTemplateServiceImpl) OnStart() error {
	if s.ginEngine != nil {
		return s.ReloadTemplates()
	}
	s.loaded = true
	return nil
}

func (s *liteHTMLTemplateServiceImpl) OnStop() error {
	s.loaded = false
	return nil
}

func (s *liteHTMLTemplateServiceImpl) SetGinEngine(engine *gin.Engine) {
	s.ginEngine = engine
}

func (s *liteHTMLTemplateServiceImpl) ReloadTemplates() error {
	if s.ginEngine == nil {
		return fmt.Errorf("gin engine not set")
	}

	funcMap := s.buildFuncMap()

	templ := template.New("").
		Funcs(funcMap).
		Delims(s.cfg.getLeftDelim(), s.cfg.getRightDelim())

	templ, err := templ.ParseGlob(s.cfg.getTemplatePath())
	if err != nil {
		if s.LoggerMgr != nil {
			s.LoggerMgr.Ins().Error("模板加载失败", "path", s.cfg.getTemplatePath(), "error", err)
		}
		return fmt.Errorf("解析模板失败: %w", err)
	}

	s.templates = templ
	s.ginEngine.SetHTMLTemplate(templ)
	s.loaded = true

	if s.LoggerMgr != nil {
		s.LoggerMgr.Ins().Info("模板加载成功", "path", s.cfg.getTemplatePath())
	}

	return nil
}

func (s *liteHTMLTemplateServiceImpl) buildFuncMap() template.FuncMap {
	funcMap := make(template.FuncMap)

	for name, fn := range s.funcMap {
		funcMap[name] = fn
	}

	if s.I18nSvc != nil {
		funcMap["t"] = func(lang, key string) string {
			return s.I18nSvc.T(lang, key)
		}
		funcMap["tWithData"] = func(lang, key string, data map[string]interface{}) string {
			return s.I18nSvc.TWithData(lang, key, data)
		}
	}

	return funcMap
}

func (s *liteHTMLTemplateServiceImpl) AddFunc(name string, fn interface{}) {
	s.funcMap[name] = fn
}

func (s *liteHTMLTemplateServiceImpl) AddFuncMap(funcMap template.FuncMap) {
	for name, fn := range funcMap {
		s.funcMap[name] = fn
	}
}

func (s *liteHTMLTemplateServiceImpl) Render(ctx *gin.Context, name string, data interface{}) {
	s.RenderWithCode(ctx, http.StatusOK, name, data)
}

func (s *liteHTMLTemplateServiceImpl) RenderWithCode(ctx *gin.Context, code int, name string, data interface{}) {
	if s.ginEngine == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "gin engine not set"})
		return
	}
	if !s.loaded {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "templates not loaded"})
		return
	}
	ctx.HTML(code, name, data)
}

func (s *liteHTMLTemplateServiceImpl) GetTemplateNames() []string {
	if s.templates == nil {
		return nil
	}

	var names []string
	for _, t := range s.templates.Templates() {
		if t.Name() != "" {
			names = append(names, t.Name())
		}
	}
	return names
}

func (s *liteHTMLTemplateServiceImpl) HasTemplate(name string) bool {
	if s.templates == nil {
		return false
	}

	return s.templates.Lookup(name) != nil
}

var _ ILiteHTMLTemplateService = (*liteHTMLTemplateServiceImpl)(nil)
var _ common.IBaseService = (*liteHTMLTemplateServiceImpl)(nil)

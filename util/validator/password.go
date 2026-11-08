package validator

import (
	"fmt"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// PasswordConfig 密码验证配置
type PasswordConfig struct {
	// MinLength 密码最小长度（默认 12）
	MinLength int
	// MaxLength 密码最大长度（默认 128）
	MaxLength int
	// RequireUpper 是否要求大写字母（默认 true）
	RequireUpper bool
	// RequireLower 是否要求小写字母（默认 true）
	RequireLower bool
	// RequireNumber 是否要求数字（默认 true）
	RequireNumber bool
	// RequireSpecial 是否要求特殊字符（默认 true）
	RequireSpecial bool
}

// DefaultPasswordConfig 返回默认密码配置
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		MinLength:      12,
		MaxLength:      128,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}
}

// passwordValidator 密码验证器实例
type passwordValidator struct {
	config *PasswordConfig
}

// newPasswordValidator 创建密码验证器
func newPasswordValidator(config *PasswordConfig) *passwordValidator {
	if config == nil {
		config = DefaultPasswordConfig()
	}
	return &passwordValidator{config: config}
}

// validate 验证密码复杂度
func (v *passwordValidator) validate(password string) error {
	// 检查长度
	if len(password) < v.config.MinLength {
		return fmt.Errorf("password must be at least %d characters", v.config.MinLength)
	}
	if len(password) > v.config.MaxLength {
		return fmt.Errorf("password must be at most %d characters", v.config.MaxLength)
	}

	var (
		hasUpper   = !v.config.RequireUpper
		hasLower   = !v.config.RequireLower
		hasNumber  = !v.config.RequireNumber
		hasSpecial = !v.config.RequireSpecial
	)

	for _, char := range password {
		switch {
		case v.config.RequireUpper && unicode.IsUpper(char):
			hasUpper = true
		case v.config.RequireLower && unicode.IsLower(char):
			hasLower = true
		case v.config.RequireNumber && unicode.IsNumber(char):
			hasNumber = true
		case v.config.RequireSpecial && (unicode.IsPunct(char) || unicode.IsSymbol(char)):
			hasSpecial = true
		}
	}

	// 收集缺失的要求
	var missingReqs []string
	if !hasUpper && v.config.RequireUpper {
		missingReqs = append(missingReqs, "uppercase letter")
	}
	if !hasLower && v.config.RequireLower {
		missingReqs = append(missingReqs, "lowercase letter")
	}
	if !hasNumber && v.config.RequireNumber {
		missingReqs = append(missingReqs, "number")
	}
	if !hasSpecial && v.config.RequireSpecial {
		missingReqs = append(missingReqs, "special character")
	}

	if len(missingReqs) > 0 {
		return fmt.Errorf("password must contain at least one: %s",
			formatRequirements(missingReqs))
	}

	return nil
}

// formatRequirements 格式化要求列表
func formatRequirements(reqs []string) string {
	result := ""
	for i, req := range reqs {
		if i > 0 {
			result += ", "
		}
		result += req
	}
	return result
}

// ValidateComplexPassword 使用默认配置验证密码复杂度
// 这是一个便捷函数，使用默认配置进行验证
func ValidateComplexPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	validator := newPasswordValidator(DefaultPasswordConfig())
	return validator.validate(password) == nil
}

// ValidatePassword 使用自定义配置验证密码并返回详细错误信息
// 这个函数可以在服务层使用，提供更友好的错误提示
func ValidatePassword(password string, config *PasswordConfig) error {
	validator := newPasswordValidator(config)
	return validator.validate(password)
}

// RegisterPasswordValidation 注册密码复杂度验证器到 validator 实例
// 使用默认配置
func RegisterPasswordValidation(v *DefaultValidator) error {
	return v.RegisterValidation("complexPassword", ValidateComplexPassword)
}

// RegisterPasswordValidationWithConfig 使用自定义配置注册密码验证器
func RegisterPasswordValidationWithConfig(v *DefaultValidator, config *PasswordConfig) error {
	pwValidator := newPasswordValidator(config)
	return v.RegisterValidation("complexPassword", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		return pwValidator.validate(password) == nil
	})
}

// GetPasswordRequirements 获取密码要求说明（用于错误提示）
// 使用默认配置
func GetPasswordRequirements() string {
	return GetPasswordRequirementsWithConfig(DefaultPasswordConfig())
}

// GetPasswordRequirementsWithConfig 使用自定义配置获取密码要求说明
func GetPasswordRequirementsWithConfig(config *PasswordConfig) string {
	var reqs []string

	if config.MinLength > 0 {
		reqs = append(reqs, fmt.Sprintf("at least %d characters", config.MinLength))
	}
	if config.RequireUpper {
		reqs = append(reqs, "uppercase letter")
	}
	if config.RequireLower {
		reqs = append(reqs, "lowercase letter")
	}
	if config.RequireNumber {
		reqs = append(reqs, "number")
	}
	if config.RequireSpecial {
		reqs = append(reqs, "special character")
	}

	return fmt.Sprintf("Password must contain: %s", formatRequirements(reqs))
}

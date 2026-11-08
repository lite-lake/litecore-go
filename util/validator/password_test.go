package validator

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestDefaultPasswordConfig 测试默认配置
func TestDefaultPasswordConfig(t *testing.T) {
	config := DefaultPasswordConfig()

	assert.Equal(t, 12, config.MinLength, "最小密码长度应为 12")
	assert.Equal(t, 128, config.MaxLength, "最大密码长度应为 128")
	assert.True(t, config.RequireUpper, "RequireUpper 应为 true")
	assert.True(t, config.RequireLower, "RequireLower 应为 true")
	assert.True(t, config.RequireNumber, "RequireNumber 应为 true")
	assert.True(t, config.RequireSpecial, "RequireSpecial 应为 true")
}

// TestValidatePassword 测试密码验证函数
func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid password",
			password:    "SecureP@ssw0rd123",
			expectError: false,
		},
		{
			name:        "valid password with special chars at end",
			password:    "MySecurePass123!@#",
			expectError: false,
		},
		{
			name:        "too short",
			password:    "Short1!",
			expectError: true,
			errorMsg:    "must be at least 12 characters",
		},
		{
			name:        "missing uppercase",
			password:    "lowercase123!@#",
			expectError: true,
			errorMsg:    "uppercase letter",
		},
		{
			name:        "missing lowercase",
			password:    "UPPERCASE123!@#",
			expectError: true,
			errorMsg:    "lowercase letter",
		},
		{
			name:        "missing number",
			password:    "NoNumbersHere!@#",
			expectError: true,
			errorMsg:    "number",
		},
		{
			name:        "missing special character",
			password:    "NoSpecialChar123",
			expectError: true,
			errorMsg:    "special character",
		},
		{
			name:        "missing multiple requirements",
			password:    "nouppercase1!",
			expectError: true,
			errorMsg:    "uppercase letter",
		},
		{
			name:        "only numbers",
			password:    "123456789012",
			expectError: true,
			errorMsg:    "uppercase",
		},
		{
			name:        "empty string",
			password:    "",
			expectError: true,
			errorMsg:    "must be at least 12 characters",
		},
		{
			name:        "exactly 12 chars with all requirements",
			password:    "Abcdefg1!xyz",
			expectError: false,
		},
		{
			name:        "very long password",
			password:    "ThisIsAVeryLongPassword123!@#ButItShouldStillWork",
			expectError: false,
		},
		{
			name:        "contains only letters and numbers",
			password:    "NoSpecialChars123",
			expectError: true,
			errorMsg:    "special character",
		},
		{
			name:        "all special characters no letters",
			password:    "!@#$%^&*()123",
			expectError: true,
			errorMsg:    "uppercase",
		},
	}

	config := DefaultPasswordConfig()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, config)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidatePasswordWithCustomConfig 测试自定义配置
func TestValidatePasswordWithCustomConfig(t *testing.T) {
	t.Run("only requires length and number", func(t *testing.T) {
		config := &PasswordConfig{
			MinLength:      8,
			MaxLength:      64,
			RequireUpper:   false,
			RequireLower:   false,
			RequireNumber:  true,
			RequireSpecial: false,
		}

		tests := []struct {
			name        string
			password    string
			expectError bool
		}{
			{
				name:        "valid - lowercase and number",
				password:    "password123",
				expectError: false,
			},
			{
				name:        "valid - uppercase and number",
				password:    "PASSWORD123",
				expectError: false,
			},
			{
				name:        "valid - mixed and number",
				password:    "Password123",
				expectError: false,
			},
			{
				name:        "invalid - no number",
				password:    "password",
				expectError: true,
			},
			{
				name:        "invalid - too short",
				password:    "pass1",
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidatePassword(tt.password, config)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("requires uppercase, lowercase and special only", func(t *testing.T) {
		config := &PasswordConfig{
			MinLength:      8,
			MaxLength:      64,
			RequireUpper:   true,
			RequireLower:   true,
			RequireNumber:  false,
			RequireSpecial: true,
		}

		tests := []struct {
			name        string
			password    string
			expectError bool
		}{
			{
				name:        "valid - has upper, lower and special",
				password:    "Password!@#",
				expectError: false,
			},
			{
				name:        "invalid - missing special",
				password:    "Password123",
				expectError: true,
			},
			{
				name:        "invalid - missing uppercase",
				password:    "password!@#",
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidatePassword(tt.password, config)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("minimal requirements", func(t *testing.T) {
		config := &PasswordConfig{
			MinLength:      1,
			MaxLength:      100,
			RequireUpper:   false,
			RequireLower:   false,
			RequireNumber:  false,
			RequireSpecial: false,
		}

		// 任何非空密码都应该通过
		err := ValidatePassword("a", config)
		assert.NoError(t, err)
	})
}

// TestGetPasswordRequirements 测试获取密码要求
func TestGetPasswordRequirements(t *testing.T) {
	t.Run("default configmgr", func(t *testing.T) {
		requirements := GetPasswordRequirements()
		assert.Contains(t, requirements, "12 characters")
		assert.Contains(t, requirements, "uppercase")
		assert.Contains(t, requirements, "lowercase")
		assert.Contains(t, requirements, "number")
		assert.Contains(t, requirements, "special character")
	})

	t.Run("custom configmgr", func(t *testing.T) {
		config := &PasswordConfig{
			MinLength:      8,
			RequireUpper:   true,
			RequireLower:   true,
			RequireNumber:  false,
			RequireSpecial: false,
		}

		requirements := GetPasswordRequirementsWithConfig(config)
		assert.Contains(t, requirements, "8 characters")
		assert.Contains(t, requirements, "uppercase")
		assert.Contains(t, requirements, "lowercase")
		assert.NotContains(t, requirements, "number")
		assert.NotContains(t, requirements, "special character")
	})
}

// TestPasswordValidationIntegration 测试密码验证的完整流程
func TestPasswordValidationIntegration(t *testing.T) {
	t.Run("password_validation_requirements", func(t *testing.T) {
		reqs := GetPasswordRequirements()
		assert.NotEmpty(t, reqs)
		assert.Contains(t, reqs, "12")
		assert.Contains(t, reqs, "uppercase")
		assert.Contains(t, reqs, "lowercase")
		assert.Contains(t, reqs, "number")
		assert.Contains(t, reqs, "special")
	})

	t.Run("accepts_secure_passwords", func(t *testing.T) {
		config := DefaultPasswordConfig()
		securePasswords := []string{
			"SecureP@ssw0rd123",
			"MySecurePass123!@#",
			"Abcdefg1!xyz",
			"Complex$Passw0rd2024!",
			"Str0ng!P@ssw0rd#123",
		}

		for _, pwd := range securePasswords {
			t.Run(pwd, func(t *testing.T) {
				err := ValidatePassword(pwd, config)
				assert.NoError(t, err, "Password should be valid: %s", pwd)
			})
		}
	})

	t.Run("rejects_weak_passwords", func(t *testing.T) {
		config := DefaultPasswordConfig()
		weakPasswords := []struct {
			password string
			reason   string
		}{
			{"short", "too short"},
			{"alllowercase", "no uppercase"},
			{"ALLUPPERCASE1!", "no lowercase"},
			{"NoNumbers!", "no numbers"},
			{"NoSpecialChars123", "no special chars"},
			{"123456789012", "only numbers"},
			{"!@#$%^&*()_+", "only special chars"},
			{"", "empty"},
		}

		for _, wp := range weakPasswords {
			t.Run(wp.reason, func(t *testing.T) {
				err := ValidatePassword(wp.password, config)
				assert.Error(t, err, "Password should be rejected: %s (%s)", wp.password, wp.reason)
			})
		}
	})

	t.Run("password_constants", func(t *testing.T) {
		config := DefaultPasswordConfig()
		assert.Equal(t, 12, config.MinLength, "Minimum password length should be 12")
		assert.Equal(t, 128, config.MaxLength, "Maximum password length should be 128")

		t.Run("exactly_min_length", func(t *testing.T) {
			pwd := "Abcdefg1!xyz"
			err := ValidatePassword(pwd, config)
			assert.NoError(t, err)
			assert.Equal(t, 12, len(pwd))
		})

		t.Run("one_less_than_min", func(t *testing.T) {
			pwd := "Abcdefg1!xy"
			err := ValidatePassword(pwd, config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "12")
		})

		t.Run("exactly_max_length", func(t *testing.T) {
			base := "Aa1!Aa1@Aa1#"
			pwd := ""
			for len(pwd) < 128 {
				pwd += base
			}
			pwd = pwd[:128]

			err := ValidatePassword(pwd, config)
			assert.NoError(t, err)
			assert.Equal(t, 128, len(pwd))
		})

		t.Run("one_more_than_max", func(t *testing.T) {
			pwd := ""
			for i := 0; i < 129; i++ {
				pwd += "A"
			}
			pwd += "1!"

			err := ValidatePassword(pwd, config)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "128")
		})
	})

	t.Run("realistic_password_scenarios", func(t *testing.T) {
		config := DefaultPasswordConfig()
		scenarios := []struct {
			name     string
			password string
			valid    bool
		}{
			{
				name:     "phrase_with_numbers_and_symbols",
				password: "Correct-Horse-Battery-Staple-42!",
				valid:    true,
			},
			{
				name:     "random_complex_password",
				password: "xK9$mP2@nL5#qR8",
				valid:    true,
			},
			{
				name:     "keyboard_pattern_weak",
				password: "Qwerty123!@#",
				valid:    true,
			},
			{
				name:     "common_weak_password_1",
				password: "Password123!",
				valid:    true,
			},
			{
				name:     "all_numbers_plus_special",
				password: "123456789012!@",
				valid:    false,
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				err := ValidatePassword(scenario.password, config)
				if scenario.valid {
					assert.NoError(t, err, "Password should be valid: %s", scenario.password)
				} else {
					assert.Error(t, err, "Password should be invalid: %s", scenario.password)
				}
			})
		}
	})

	t.Run("error_messages_are_helpful", func(t *testing.T) {
		config := DefaultPasswordConfig()
		testCases := []struct {
			password string
			expected string
		}{
			{"short", "12"},
			{"nouppercase1!", "uppercase"},
			{"NOLOWERCASE1!", "lowercase"},
			{"NoNumbersHere!", "number"},
			{"NoSpecialChars123", "special"},
		}

		for _, tc := range testCases {
			t.Run(tc.password, func(t *testing.T) {
				err := ValidatePassword(tc.password, config)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expected,
					"Error message should mention '%s' for password '%s'", tc.expected, tc.password)
			})
		}
	})
}

// TestValidateComplexPassword 测试 ValidateComplexPassword 函数
func TestValidateComplexPassword(t *testing.T) {
	v := NewDefaultValidator()

	// 先注册密码验证器
	err := RegisterPasswordValidation(v)
	assert.NoError(t, err)

	type TestRequest struct {
		Password string `json:"password" validate:"complexPassword"`
	}

	tests := []struct {
		name    string
		reqBody string
		wantErr bool
	}{
		{
			name:    "Valid complex password",
			reqBody: `{"password":"SecureP@ssw0rd123"}`,
			wantErr: false,
		},
		{
			name:    "Too short",
			reqBody: `{"password":"Short1!"}`,
			wantErr: true,
		},
		{
			name:    "Missing uppercase",
			reqBody: `{"password":"lowercase123!@#"}`,
			wantErr: true,
		},
		{
			name:    "Missing lowercase",
			reqBody: `{"password":"UPPERCASE123!@#"}`,
			wantErr: true,
		},
		{
			name:    "Missing number",
			reqBody: `{"password":"NoNumbersHere!@#"}`,
			wantErr: true,
		},
		{
			name:    "Missing special character",
			reqBody: `{"password":"NoSpecialChars123"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var req TestRequest
			err := v.Validate(c, &req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRegisterPasswordValidation 测试 RegisterPasswordValidation 函数
func TestRegisterPasswordValidation(t *testing.T) {
	v := NewDefaultValidator()

	// 注册密码验证器
	err := RegisterPasswordValidation(v)
	assert.NoError(t, err)

	// 验证注册成功
	type TestRequest struct {
		Password string `json:"password" validate:"complexPassword"`
	}

	reqBody := `{"password":"SecureP@ssw0rd123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	var req TestRequest
	err = v.Validate(c, &req)
	assert.NoError(t, err)
}

// TestRegisterPasswordValidationWithConfig 测试 RegisterPasswordValidationWithConfig 函数
func TestRegisterPasswordValidationWithConfig(t *testing.T) {
	v := NewDefaultValidator()

	// 使用自定义配置注册密码验证器
	config := &PasswordConfig{
		MinLength:      8,
		MaxLength:      64,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: false,
	}

	err := RegisterPasswordValidationWithConfig(v, config)
	assert.NoError(t, err)

	// 验证注册成功 - 这个密码应该通过（因为不需要特殊字符）
	type TestRequest struct {
		Password string `json:"password" validate:"complexPassword"`
	}

	reqBody := `{"password":"Password123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	var req TestRequest
	err = v.Validate(c, &req)
	assert.NoError(t, err)

	// 这个密码应该失败（没有大写）
	reqBody2 := `{"password":"password123"}`
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody2))
	c2.Request.Header.Set("Content-Type", "application/json")

	var req2 TestRequest
	err = v.Validate(c2, &req2)
	assert.Error(t, err)
}

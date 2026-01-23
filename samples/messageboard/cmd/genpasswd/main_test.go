package main

import (
	"strings"
	"testing"

	"github.com/lite-lake/litecore-go/util/hash"
)

// TestGeneratePassword 测试密码生成功能
func TestGeneratePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "简单密码",
			password: "admin123",
			wantErr:  false,
		},
		{
			name:     "复杂密码",
			password: "MyP@ssw0rd!2024#",
			wantErr:  false,
		},
		{
			name:     "中文字符密码",
			password: "密码123",
			wantErr:  false,
		},
		{
			name:     "空密码",
			password: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := hash.Hash.BcryptHash(tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("BcryptHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if hashedPassword == "" {
					t.Error("BcryptHash() 返回空字符串")
				}

				if !strings.HasPrefix(hashedPassword, "$2a$10$") && !strings.HasPrefix(hashedPassword, "$2b$10$") {
					t.Errorf("BcryptHash() 返回的哈希值格式不正确: %s", hashedPassword)
				}

				if hash.Hash.BcryptVerify(tt.password, hashedPassword) != true {
					t.Error("BcryptVerify() 验证失败，生成的哈希值无法验证")
				}
			}
		})
	}
}

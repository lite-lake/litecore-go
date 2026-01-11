package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
	"time"
)

// Test helper functions

// generateTestRSAKey 生成测试用的RSA密钥对
func generateTestRSAKey() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// generateTestECDSAKey 生成测试用的ECDSA密钥对
func generateTestECDSAKey() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// =========================================
// Test: JWT Generation - HS256
// =========================================

func TestGenerateHS256Token(t *testing.T) {
	secretKey := []byte("test-secret-key-for-hs256")

	tests := []struct {
		name    string
		claims  ILiteUtilJWTClaims
		wantErr bool
	}{
		{
			name: "valid StandardClaims",
			claims: &StandardClaims{
				Issuer:    "test-issuer",
				Subject:   "test-subject",
				Audience:  []string{"test-audience"},
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				IssuedAt:  time.Now().Unix(),
			},
			wantErr: false,
		},
		{
			name: "valid MapClaims",
			claims: MapClaims{
				"iss":          "test-issuer",
				"sub":          "test-subject",
				"aud":          "test-audience",
				"exp":          float64(time.Now().Add(time.Hour).Unix()),
				"iat":          float64(time.Now().Unix()),
				"custom_field": "custom-value",
			},
			wantErr: false,
		},
		{
			name:    "empty claims",
			claims:  MapClaims{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := JWT.GenerateHS256Token(tt.claims, secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateHS256Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if token == "" {
					t.Error("GenerateHS256Token() returned empty token")
				}

				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					t.Errorf("GenerateHS256Token() invalid token format, got %d parts", len(parts))
				}
			}
		})
	}
}

// =========================================
// Test: JWT Generation - HS512
// =========================================

func TestGenerateHS512Token(t *testing.T) {
	secretKey := []byte("test-secret-key-for-hs512")

	tests := []struct {
		name    string
		claims  ILiteUtilJWTClaims
		wantErr bool
	}{
		{
			name: "valid StandardClaims",
			claims: &StandardClaims{
				Issuer:    "test-issuer",
				Subject:   "test-subject",
				Audience:  []string{"test-audience"},
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				IssuedAt:  time.Now().Unix(),
			},
			wantErr: false,
		},
		{
			name: "valid MapClaims with custom fields",
			claims: MapClaims{
				"iss": "test-issuer",
				"sub": "test-subject",
				"exp": float64(time.Now().Add(time.Hour).Unix()),
				"custom_data": map[string]interface{}{
					"key1": "value1",
					"key2": 12345,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := JWT.GenerateHS512Token(tt.claims, secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateHS512Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token == "" {
				t.Error("GenerateHS512Token() returned empty token")
			}
		})
	}
}

// =========================================
// Test: JWT Generation - RS256
// =========================================

func TestGenerateRS256Token(t *testing.T) {
	privateKey, _, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	tests := []struct {
		name    string
		claims  ILiteUtilJWTClaims
		wantErr bool
	}{
		{
			name: "valid StandardClaims",
			claims: &StandardClaims{
				Issuer:    "test-issuer",
				Subject:   "test-subject",
				Audience:  []string{"test-audience"},
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				IssuedAt:  time.Now().Unix(),
			},
			wantErr: false,
		},
		{
			name: "valid MapClaims",
			claims: MapClaims{
				"iss": "test-issuer",
				"sub": "test-subject",
				"aud": []string{"audience1", "audience2"},
				"exp": float64(time.Now().Add(time.Hour).Unix()),
			},
			wantErr: false,
		},
		{
			name:    "nil private key",
			claims:  &StandardClaims{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var token string
			var err error

			if tt.name == "nil private key" {
				token, err = JWT.GenerateRS256Token(tt.claims, nil)
			} else {
				token, err = JWT.GenerateRS256Token(tt.claims, privateKey)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRS256Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token == "" {
				t.Error("GenerateRS256Token() returned empty token")
			}
		})
	}
}

// =========================================
// Test: JWT Generation - ES256
// =========================================

func TestGenerateES256Token(t *testing.T) {
	privateKey, _, err := generateTestECDSAKey()
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	tests := []struct {
		name    string
		claims  ILiteUtilJWTClaims
		wantErr bool
	}{
		{
			name: "valid StandardClaims",
			claims: &StandardClaims{
				Issuer:    "test-issuer",
				Subject:   "test-subject",
				Audience:  []string{"test-audience"},
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				IssuedAt:  time.Now().Unix(),
			},
			wantErr: false,
		},
		{
			name: "valid MapClaims",
			claims: MapClaims{
				"iss": "test-issuer",
				"sub": "test-subject",
				"aud": "test-audience",
				"exp": float64(time.Now().Add(time.Hour).Unix()),
			},
			wantErr: false,
		},
		{
			name:    "nil private key",
			claims:  &StandardClaims{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var token string
			var err error

			if tt.name == "nil private key" {
				token, err = JWT.GenerateES256Token(tt.claims, nil)
			} else {
				token, err = JWT.GenerateES256Token(tt.claims, privateKey)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateES256Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && token == "" {
				t.Error("GenerateES256Token() returned empty token")
			}
		})
	}
}

// =========================================
// Test: JWT Parse - HS256
// =========================================

func TestParseHS256Token(t *testing.T) {
	secretKey := []byte("test-secret-key")

	// Create a valid token first
	claims := MapClaims{
		"iss": "test-issuer",
		"sub": "test-subject",
		"aud": "test-audience",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"iat": float64(time.Now().Unix()),
	}
	token, _ := JWT.GenerateHS256Token(claims, secretKey)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   token,
			wantErr: false,
		},
		{
			name:    "invalid format - missing parts",
			token:   "invalid.token",
			wantErr: true,
		},
		{
			name:    "invalid format - too many parts",
			token:   "a.b.c.d",
			wantErr: true,
		},
		{
			name:    "invalid signature - wrong key",
			token:   token,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var key []byte
			if tt.name == "invalid signature - wrong key" {
				key = []byte("wrong-secret-key")
			} else {
				key = secretKey
			}

			parsedClaims, err := JWT.ParseHS256Token(tt.token, key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHS256Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if parsedClaims == nil {
					t.Error("ParseHS256Token() returned nil claims")
				}
				if parsedClaims.GetIssuer() != "test-issuer" {
					t.Errorf("ParseHS256Token() issuer = %v, want %v", parsedClaims.GetIssuer(), "test-issuer")
				}
			}
		})
	}
}

// =========================================
// Test: JWT Parse - HS512
// =========================================

func TestParseHS512Token(t *testing.T) {
	secretKey := []byte("test-secret-key")

	claims := MapClaims{
		"iss": "test-issuer",
		"sub": "test-subject",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}
	token, _ := JWT.GenerateHS512Token(claims, secretKey)

	tests := []struct {
		name    string
		token   string
		key     []byte
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   token,
			key:     secretKey,
			wantErr: false,
		},
		{
			name:    "wrong secret key",
			token:   token,
			key:     []byte("wrong-key"),
			wantErr: true,
		},
		{
			name:    "invalid token format",
			token:   "invalid",
			key:     secretKey,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedClaims, err := JWT.ParseHS512Token(tt.token, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHS512Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && parsedClaims == nil {
				t.Error("ParseHS512Token() returned nil claims")
			}
		})
	}
}

// =========================================
// Test: JWT Parse - RS256
// =========================================

func TestParseRS256Token(t *testing.T) {
	privateKey, publicKey, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	claims := MapClaims{
		"iss": "test-issuer",
		"sub": "test-subject",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}
	token, _ := JWT.GenerateRS256Token(claims, privateKey)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   token,
			wantErr: false,
		},
		{
			name:    "invalid token format",
			token:   "invalid.token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedClaims, err := JWT.ParseRS256Token(tt.token, publicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRS256Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if parsedClaims == nil {
					t.Error("ParseRS256Token() returned nil claims")
				}
				if parsedClaims.GetSubject() != "test-subject" {
					t.Errorf("ParseRS256Token() subject = %v, want %v", parsedClaims.GetSubject(), "test-subject")
				}
			}
		})
	}
}

// =========================================
// Test: JWT Parse - ES256
// =========================================

func TestParseES256Token(t *testing.T) {
	privateKey, publicKey, err := generateTestECDSAKey()
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	claims := MapClaims{
		"iss": "test-issuer",
		"sub": "test-subject",
		"aud": "test-audience",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}
	token, _ := JWT.GenerateES256Token(claims, privateKey)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   token,
			wantErr: false,
		},
		{
			name:    "invalid token format",
			token:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedClaims, err := JWT.ParseES256Token(tt.token, publicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseES256Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && parsedClaims == nil {
				t.Error("ParseES256Token() returned nil claims")
			}
		})
	}
}

// =========================================
// Test: JWT Round-trip (Generate and Parse)
// =========================================

func TestJWTRoundTrip_HS256(t *testing.T) {
	secretKey := []byte("test-secret-key")

	originalClaims := MapClaims{
		"iss":           "test-issuer",
		"sub":           "test-subject",
		"aud":           []string{"audience1", "audience2"},
		"exp":           float64(time.Now().Add(time.Hour).Unix()),
		"iat":           float64(time.Now().Unix()),
		"custom_string": "hello",
		"custom_number": 12345,
		"custom_bool":   true,
	}

	token, err := JWT.GenerateHS256Token(originalClaims, secretKey)
	if err != nil {
		t.Fatalf("GenerateHS256Token() error = %v", err)
	}

	parsedClaims, err := JWT.ParseHS256Token(token, secretKey)
	if err != nil {
		t.Fatalf("ParseHS256Token() error = %v", err)
	}

	// Verify claims match
	if parsedClaims.GetIssuer() != originalClaims.GetIssuer() {
		t.Errorf("Issuer mismatch: got %v, want %v", parsedClaims.GetIssuer(), originalClaims.GetIssuer())
	}
	if parsedClaims.GetSubject() != originalClaims.GetSubject() {
		t.Errorf("Subject mismatch: got %v, want %v", parsedClaims.GetSubject(), originalClaims.GetSubject())
	}
	if parsedClaims["custom_string"] != "hello" {
		t.Errorf("custom_string mismatch: got %v, want %v", parsedClaims["custom_string"], "hello")
	}
	// JSON unmarshaling converts numbers to float64
	if parsedClaims["custom_number"] != float64(12345) {
		t.Errorf("custom_number mismatch: got %v (type %T), want 12345", parsedClaims["custom_number"], parsedClaims["custom_number"])
	}
}

func TestJWTRoundTrip_RS256(t *testing.T) {
	privateKey, publicKey, err := generateTestRSAKey()
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	originalClaims := &StandardClaims{
		Issuer:    "test-issuer",
		Subject:   "test-subject",
		Audience:  []string{"test-audience"},
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		IssuedAt:  time.Now().Unix(),
		ID:        "test-jti",
	}

	token, err := JWT.GenerateRS256Token(originalClaims, privateKey)
	if err != nil {
		t.Fatalf("GenerateRS256Token() error = %v", err)
	}

	parsedClaims, err := JWT.ParseRS256Token(token, publicKey)
	if err != nil {
		t.Fatalf("ParseRS256Token() error = %v", err)
	}

	// Verify claims match
	if parsedClaims.GetIssuer() != originalClaims.GetIssuer() {
		t.Errorf("Issuer mismatch: got %v, want %v", parsedClaims.GetIssuer(), originalClaims.GetIssuer())
	}
	if parsedClaims.GetSubject() != originalClaims.GetSubject() {
		t.Errorf("Subject mismatch: got %v, want %v", parsedClaims.GetSubject(), originalClaims.GetSubject())
	}
	if parsedClaims["jti"] != "test-jti" {
		t.Errorf("jti mismatch: got %v, want %v", parsedClaims["jti"], "test-jti")
	}
}

func TestJWTRoundTrip_ES256(t *testing.T) {
	privateKey, publicKey, err := generateTestECDSAKey()
	if err != nil {
		t.Fatalf("Failed to generate ECDSA key: %v", err)
	}

	originalClaims := MapClaims{
		"iss":     "ecdsa-issuer",
		"sub":     "ecdsa-subject",
		"exp":     float64(time.Now().Add(time.Hour).Unix()),
		"user_id": float64(12345),
		"role":    "admin",
	}

	token, err := JWT.GenerateES256Token(originalClaims, privateKey)
	if err != nil {
		t.Fatalf("GenerateES256Token() error = %v", err)
	}

	parsedClaims, err := JWT.ParseES256Token(token, publicKey)
	if err != nil {
		t.Fatalf("ParseES256Token() error = %v", err)
	}

	// Verify claims match
	if parsedClaims.GetIssuer() != "ecdsa-issuer" {
		t.Errorf("Issuer mismatch: got %v, want %v", parsedClaims.GetIssuer(), "ecdsa-issuer")
	}
	if parsedClaims["role"] != "admin" {
		t.Errorf("role mismatch: got %v, want %v", parsedClaims["role"], "admin")
	}
}

// =========================================
// Test: StandardClaims Methods
// =========================================

func TestStandardClaims_Getters(t *testing.T) {
	now := time.Now()
	exp := now.Add(time.Hour)
	nbf := now.Add(-time.Minute)

	claims := &StandardClaims{
		Audience:  []string{"aud1", "aud2"},
		ExpiresAt: exp.Unix(),
		ID:        "test-id",
		IssuedAt:  now.Unix(),
		Issuer:    "test-issuer",
		NotBefore: nbf.Unix(),
		Subject:   "test-subject",
	}

	t.Run("GetExpiresAt", func(t *testing.T) {
		got := claims.GetExpiresAt()
		if got == nil {
			t.Error("GetExpiresAt() returned nil")
		} else if got.Unix() != exp.Unix() {
			t.Errorf("GetExpiresAt() = %v, want %v", got, exp)
		}
	})

	t.Run("GetIssuedAt", func(t *testing.T) {
		got := claims.GetIssuedAt()
		if got == nil {
			t.Error("GetIssuedAt() returned nil")
		} else if got.Unix() != now.Unix() {
			t.Errorf("GetIssuedAt() = %v, want %v", got, now)
		}
	})

	t.Run("GetNotBefore", func(t *testing.T) {
		got := claims.GetNotBefore()
		if got == nil {
			t.Error("GetNotBefore() returned nil")
		} else if got.Unix() != nbf.Unix() {
			t.Errorf("GetNotBefore() = %v, want %v", got, nbf)
		}
	})

	t.Run("GetIssuer", func(t *testing.T) {
		if got := claims.GetIssuer(); got != "test-issuer" {
			t.Errorf("GetIssuer() = %v, want %v", got, "test-issuer")
		}
	})

	t.Run("GetSubject", func(t *testing.T) {
		if got := claims.GetSubject(); got != "test-subject" {
			t.Errorf("GetSubject() = %v, want %v", got, "test-subject")
		}
	})

	t.Run("GetAudience", func(t *testing.T) {
		got := claims.GetAudience()
		if len(got) != 2 || got[0] != "aud1" || got[1] != "aud2" {
			t.Errorf("GetAudience() = %v, want [aud1, aud2]", got)
		}
	})

	t.Run("GetCustomClaims", func(t *testing.T) {
		got := claims.GetCustomClaims()
		if got == nil || len(got) != 0 {
			t.Errorf("GetCustomClaims() should return empty map, got %v", got)
		}
	})
}

func TestStandardClaims_NilValues(t *testing.T) {
	claims := &StandardClaims{}

	t.Run("GetExpiresAt when zero", func(t *testing.T) {
		if got := claims.GetExpiresAt(); got != nil {
			t.Errorf("GetExpiresAt() should return nil when ExpiresAt is 0, got %v", got)
		}
	})

	t.Run("GetIssuedAt when zero", func(t *testing.T) {
		if got := claims.GetIssuedAt(); got != nil {
			t.Errorf("GetIssuedAt() should return nil when IssuedAt is 0, got %v", got)
		}
	})

	t.Run("GetNotBefore when zero", func(t *testing.T) {
		if got := claims.GetNotBefore(); got != nil {
			t.Errorf("GetNotBefore() should return nil when NotBefore is 0, got %v", got)
		}
	})
}

// =========================================
// Test: MapClaims Methods
// =========================================

func TestMapClaims_Getters(t *testing.T) {
	now := time.Now()

	claims := MapClaims{
		"iss":           "test-issuer",
		"sub":           "test-subject",
		"aud":           []string{"aud1", "aud2"},
		"exp":           float64(now.Add(time.Hour).Unix()),
		"iat":           float64(now.Unix()),
		"nbf":           float64(now.Add(-time.Minute).Unix()),
		"custom_field":  "custom-value",
		"custom_number": 12345,
	}

	t.Run("GetExpiresAt", func(t *testing.T) {
		got := claims.GetExpiresAt()
		if got == nil {
			t.Error("GetExpiresAt() returned nil")
		}
	})

	t.Run("GetIssuedAt", func(t *testing.T) {
		got := claims.GetIssuedAt()
		if got == nil {
			t.Error("GetIssuedAt() returned nil")
		}
	})

	t.Run("GetNotBefore", func(t *testing.T) {
		got := claims.GetNotBefore()
		if got == nil {
			t.Error("GetNotBefore() returned nil")
		}
	})

	t.Run("GetIssuer", func(t *testing.T) {
		if got := claims.GetIssuer(); got != "test-issuer" {
			t.Errorf("GetIssuer() = %v, want %v", got, "test-issuer")
		}
	})

	t.Run("GetSubject", func(t *testing.T) {
		if got := claims.GetSubject(); got != "test-subject" {
			t.Errorf("GetSubject() = %v, want %v", got, "test-subject")
		}
	})

	t.Run("GetAudience with array", func(t *testing.T) {
		got := claims.GetAudience()
		if len(got) != 2 || got[0] != "aud1" || got[1] != "aud2" {
			t.Errorf("GetAudience() = %v, want [aud1, aud2]", got)
		}
	})

	t.Run("GetCustomClaims", func(t *testing.T) {
		got := claims.GetCustomClaims()
		if len(got) != 2 {
			t.Errorf("GetCustomClaims() should return 2 custom fields, got %d", len(got))
		}
		if got["custom_field"] != "custom-value" {
			t.Errorf("GetCustomClaims()[custom_field] = %v, want %v", got["custom_field"], "custom-value")
		}
	})
}

func TestMapClaims_AudienceVariations(t *testing.T) {
	t.Run("Audience as string", func(t *testing.T) {
		claims := MapClaims{"aud": "single-audience"}
		got := claims.GetAudience()
		if len(got) != 1 || got[0] != "single-audience" {
			t.Errorf("GetAudience() = %v, want [single-audience]", got)
		}
	})

	t.Run("Audience as []string", func(t *testing.T) {
		claims := MapClaims{"aud": []string{"aud1", "aud2"}}
		got := claims.GetAudience()
		if len(got) != 2 || got[0] != "aud1" || got[1] != "aud2" {
			t.Errorf("GetAudience() = %v, want [aud1, aud2]", got)
		}
	})

	t.Run("Audience as []interface{}", func(t *testing.T) {
		claims := MapClaims{"aud": []interface{}{"aud1", "aud2", "aud3"}}
		got := claims.GetAudience()
		if len(got) != 3 || got[0] != "aud1" || got[1] != "aud2" || got[2] != "aud3" {
			t.Errorf("GetAudience() = %v, want [aud1, aud2, aud3]", got)
		}
	})

	t.Run("No audience", func(t *testing.T) {
		claims := MapClaims{}
		got := claims.GetAudience()
		if len(got) != 0 {
			t.Errorf("GetAudience() should return empty slice, got %v", got)
		}
	})
}

func TestMapClaims_NilValues(t *testing.T) {
	claims := MapClaims{}

	t.Run("GetExpiresAt when missing", func(t *testing.T) {
		if got := claims.GetExpiresAt(); got != nil {
			t.Errorf("GetExpiresAt() should return nil when exp is missing, got %v", got)
		}
	})

	t.Run("GetIssuedAt when missing", func(t *testing.T) {
		if got := claims.GetIssuedAt(); got != nil {
			t.Errorf("GetIssuedAt() should return nil when iat is missing, got %v", got)
		}
	})

	t.Run("GetNotBefore when missing", func(t *testing.T) {
		if got := claims.GetNotBefore(); got != nil {
			t.Errorf("GetNotBefore() should return nil when nbf is missing, got %v", got)
		}
	})

	t.Run("GetIssuer when missing", func(t *testing.T) {
		if got := claims.GetIssuer(); got != "" {
			t.Errorf("GetIssuer() should return empty string when iss is missing, got %v", got)
		}
	})

	t.Run("GetSubject when missing", func(t *testing.T) {
		if got := claims.GetSubject(); got != "" {
			t.Errorf("GetSubject() should return empty string when sub is missing, got %v", got)
		}
	})
}

func TestMapClaims_SetCustomClaims(t *testing.T) {
	claims := MapClaims{}
	customClaims := map[string]interface{}{
		"key1": "value1",
		"key2": 12345,
		"key3": true,
	}

	claims.SetCustomClaims(customClaims)

	if claims["key1"] != "value1" {
		t.Errorf("claims[key1] = %v, want %v", claims["key1"], "value1")
	}
	if claims["key2"] != 12345 {
		t.Errorf("claims[key2] = %v, want %v", claims["key2"], 12345)
	}
	if claims["key3"] != true {
		t.Errorf("claims[key3] = %v, want %v", claims["key3"], true)
	}
}

// =========================================
// Test: ValidateClaims
// =========================================

func TestValidateClaims_Expiration(t *testing.T) {
	t.Run("valid token - not expired", func(t *testing.T) {
		claims := MapClaims{
			"exp": float64(time.Now().Add(time.Hour).Unix()),
		}
		err := JWT.ValidateClaims(claims)
		if err != nil {
			t.Errorf("ValidateClaims() should not error for valid token, got %v", err)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		claims := MapClaims{
			"exp": float64(time.Now().Add(-time.Hour).Unix()),
		}
		err := JWT.ValidateClaims(claims)
		if err == nil {
			t.Error("ValidateClaims() should error for expired token")
		}
		if !strings.Contains(err.Error(), "expired") {
			t.Errorf("Error should mention 'expired', got %v", err)
		}
	})

	t.Run("token without expiration", func(t *testing.T) {
		claims := MapClaims{}
		err := JWT.ValidateClaims(claims)
		if err != nil {
			t.Errorf("ValidateClaims() should not error for token without expiration, got %v", err)
		}
	})
}

func TestValidateClaims_NotBefore(t *testing.T) {
	t.Run("valid token - nbf in past", func(t *testing.T) {
		claims := MapClaims{
			"nbf": float64(time.Now().Add(-time.Hour).Unix()),
		}
		err := JWT.ValidateClaims(claims)
		if err != nil {
			t.Errorf("ValidateClaims() should not error when nbf is in past, got %v", err)
		}
	})

	t.Run("invalid token - nbf in future", func(t *testing.T) {
		claims := MapClaims{
			"nbf": float64(time.Now().Add(time.Hour).Unix()),
		}
		err := JWT.ValidateClaims(claims)
		if err == nil {
			t.Error("ValidateClaims() should error when nbf is in future")
		}
		if !strings.Contains(err.Error(), "not valid yet") {
			t.Errorf("Error should mention 'not valid yet', got %v", err)
		}
	})
}

func TestValidateClaims_Issuer(t *testing.T) {
	t.Run("valid issuer", func(t *testing.T) {
		claims := MapClaims{
			"iss": "test-issuer",
		}
		err := JWT.ValidateClaims(claims, WithIssuer("test-issuer"))
		if err != nil {
			t.Errorf("ValidateClaims() should not error for valid issuer, got %v", err)
		}
	})

	t.Run("invalid issuer", func(t *testing.T) {
		claims := MapClaims{
			"iss": "wrong-issuer",
		}
		err := JWT.ValidateClaims(claims, WithIssuer("test-issuer"))
		if err == nil {
			t.Error("ValidateClaims() should error for invalid issuer")
		}
		if !strings.Contains(err.Error(), "invalid issuer") {
			t.Errorf("Error should mention 'invalid issuer', got %v", err)
		}
	})

	t.Run("no issuer validation when not specified", func(t *testing.T) {
		claims := MapClaims{}
		err := JWT.ValidateClaims(claims)
		if err != nil {
			t.Errorf("ValidateClaims() should not error when issuer validation is not specified, got %v", err)
		}
	})
}

func TestValidateClaims_Subject(t *testing.T) {
	t.Run("valid subject", func(t *testing.T) {
		claims := MapClaims{
			"sub": "test-subject",
		}
		err := JWT.ValidateClaims(claims, WithSubject("test-subject"))
		if err != nil {
			t.Errorf("ValidateClaims() should not error for valid subject, got %v", err)
		}
	})

	t.Run("invalid subject", func(t *testing.T) {
		claims := MapClaims{
			"sub": "wrong-subject",
		}
		err := JWT.ValidateClaims(claims, WithSubject("test-subject"))
		if err == nil {
			t.Error("ValidateClaims() should error for invalid subject")
		}
		if !strings.Contains(err.Error(), "invalid subject") {
			t.Errorf("Error should mention 'invalid subject', got %v", err)
		}
	})
}

func TestValidateClaims_Audience(t *testing.T) {
	t.Run("valid audience - single", func(t *testing.T) {
		claims := MapClaims{
			"aud": "test-audience",
		}
		err := JWT.ValidateClaims(claims, WithAudience("test-audience"))
		if err != nil {
			t.Errorf("ValidateClaims() should not error for valid audience, got %v", err)
		}
	})

	t.Run("valid audience - one of multiple", func(t *testing.T) {
		claims := MapClaims{
			"aud": []string{"aud1", "aud2"},
		}
		err := JWT.ValidateClaims(claims, WithAudience("aud2", "aud3"))
		if err != nil {
			t.Errorf("ValidateClaims() should not error when one audience matches, got %v", err)
		}
	})

	t.Run("invalid audience", func(t *testing.T) {
		claims := MapClaims{
			"aud": "wrong-audience",
		}
		err := JWT.ValidateClaims(claims, WithAudience("test-audience"))
		if err == nil {
			t.Error("ValidateClaims() should error for invalid audience")
		}
		if !strings.Contains(err.Error(), "invalid audience") {
			t.Errorf("Error should mention 'invalid audience', got %v", err)
		}
	})

	t.Run("audience not in token", func(t *testing.T) {
		claims := MapClaims{}
		err := JWT.ValidateClaims(claims, WithAudience("test-audience"))
		if err == nil {
			t.Error("ValidateClaims() should error when audience is required but not in token")
		}
	})
}

func TestValidateClaims_WithCurrentTime(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)

	t.Run("token not expired at specific time", func(t *testing.T) {
		claims := MapClaims{
			"exp": float64(future.Unix()),
		}
		err := JWT.ValidateClaims(claims, WithCurrentTime(past))
		if err != nil {
			t.Errorf("ValidateClaims() should not error when token is valid at specified time, got %v", err)
		}
	})

	t.Run("token expired at specific time", func(t *testing.T) {
		claims := MapClaims{
			"exp": float64(past.Unix()),
		}
		err := JWT.ValidateClaims(claims, WithCurrentTime(future))
		if err == nil {
			t.Error("ValidateClaims() should error when token is expired at specified time")
		}
	})
}

func TestValidateClaims_CombinedValidation(t *testing.T) {
	t.Run("all validations pass", func(t *testing.T) {
		claims := MapClaims{
			"iss": "test-issuer",
			"sub": "test-subject",
			"aud": []string{"aud1", "aud2"},
			"exp": float64(time.Now().Add(time.Hour).Unix()),
			"nbf": float64(time.Now().Add(-time.Hour).Unix()),
		}
		err := JWT.ValidateClaims(claims,
			WithIssuer("test-issuer"),
			WithSubject("test-subject"),
			WithAudience("aud1", "aud2"),
		)
		if err != nil {
			t.Errorf("ValidateClaims() should not error when all validations pass, got %v", err)
		}
	})

	t.Run("multiple validation failures", func(t *testing.T) {
		claims := MapClaims{
			"iss": "wrong-issuer",
			"sub": "wrong-subject",
			"exp": float64(time.Now().Add(-time.Hour).Unix()),
		}
		err := JWT.ValidateClaims(claims,
			WithIssuer("test-issuer"),
			WithSubject("test-subject"),
		)
		if err == nil {
			t.Error("ValidateClaims() should error when validations fail")
		}
		// Should return first error encountered
	})
}

// =========================================
// Test: Convenience Methods
// =========================================

func TestNewStandardClaims(t *testing.T) {
	claims := JWT.NewStandardClaims()
	if claims == nil {
		t.Error("NewStandardClaims() returned nil")
	}
}

func TestNewMapClaims(t *testing.T) {
	claims := JWT.NewMapClaims()
	if claims == nil {
		t.Error("NewMapClaims() returned nil")
	}
	if len(claims) != 0 {
		t.Errorf("NewMapClaims() should return empty map, got %d items", len(claims))
	}
}

func TestSetExpiration(t *testing.T) {
	t.Run("SetExpiration on StandardClaims", func(t *testing.T) {
		claims := JWT.NewStandardClaims()
		duration := time.Hour

		JWT.SetExpiration(claims, duration)

		if claims.ExpiresAt == 0 {
			t.Error("SetExpiration() did not set ExpiresAt")
		}

		exp := time.Unix(claims.ExpiresAt, 0)
		expected := time.Now().Add(duration)
		diff := expected.Sub(exp)
		if diff > time.Second {
			t.Errorf("SetExpiration() expiration time off by %v", diff)
		}
	})

	t.Run("SetExpiration on MapClaims", func(t *testing.T) {
		claims := JWT.NewMapClaims()
		duration := 2 * time.Hour

		JWT.SetExpiration(claims, duration)

		if claims["exp"] == nil {
			t.Error("SetExpiration() did not set exp")
		}

		expFloat := claims["exp"].(float64)
		exp := time.Unix(int64(expFloat), 0)
		expected := time.Now().Add(duration)
		diff := expected.Sub(exp)
		if diff > time.Second {
			t.Errorf("SetExpiration() expiration time off by %v", diff)
		}
	})

	t.Run("SetExpiration with zero duration", func(t *testing.T) {
		claims := JWT.NewStandardClaims()
		JWT.SetExpiration(claims, 0)
		if claims.ExpiresAt != 0 {
			t.Error("SetExpiration() should not set ExpiresAt when duration is 0")
		}
	})

	t.Run("SetExpiration with negative duration", func(t *testing.T) {
		claims := JWT.NewStandardClaims()
		JWT.SetExpiration(claims, -time.Hour)
		if claims.ExpiresAt != 0 {
			t.Error("SetExpiration() should not set ExpiresAt when duration is negative")
		}
	})
}

func TestSetIssuedAt(t *testing.T) {
	now := time.Now()

	t.Run("SetIssuedAt on StandardClaims", func(t *testing.T) {
		claims := JWT.NewStandardClaims()
		JWT.SetIssuedAt(claims, now)

		if claims.IssuedAt == 0 {
			t.Error("SetIssuedAt() did not set IssuedAt")
		}

		iat := time.Unix(claims.IssuedAt, 0)
		diff := now.Sub(iat)
		if diff > time.Second {
			t.Errorf("SetIssuedAt() time off by %v", diff)
		}
	})

	t.Run("SetIssuedAt on MapClaims", func(t *testing.T) {
		claims := JWT.NewMapClaims()
		JWT.SetIssuedAt(claims, now)

		if claims["iat"] == nil {
			t.Error("SetIssuedAt() did not set iat")
		}

		iatFloat := claims["iat"].(float64)
		iat := time.Unix(int64(iatFloat), 0)
		diff := now.Sub(iat)
		if diff > time.Second {
			t.Errorf("SetIssuedAt() time off by %v", diff)
		}
	})
}

func TestSetNotBefore(t *testing.T) {
	future := time.Now().Add(time.Hour)

	t.Run("SetNotBefore on StandardClaims", func(t *testing.T) {
		claims := JWT.NewStandardClaims()
		JWT.SetNotBefore(claims, future)

		if claims.NotBefore == 0 {
			t.Error("SetNotBefore() did not set NotBefore")
		}

		nbf := time.Unix(claims.NotBefore, 0)
		diff := future.Sub(nbf)
		if diff > time.Second {
			t.Errorf("SetNotBefore() time off by %v", diff)
		}
	})

	t.Run("SetNotBefore on MapClaims", func(t *testing.T) {
		claims := JWT.NewMapClaims()
		JWT.SetNotBefore(claims, future)

		if claims["nbf"] == nil {
			t.Error("SetNotBefore() did not set nbf")
		}

		nbfFloat := claims["nbf"].(float64)
		nbf := time.Unix(int64(nbfFloat), 0)
		diff := future.Sub(nbf)
		if diff > time.Second {
			t.Errorf("SetNotBefore() time off by %v", diff)
		}
	})
}

func TestSetIssuer(t *testing.T) {
	t.Run("SetIssuer on StandardClaims", func(t *testing.T) {
		claims := JWT.NewStandardClaims()
		JWT.SetIssuer(claims, "test-issuer")

		if claims.Issuer != "test-issuer" {
			t.Errorf("SetIssuer() = %v, want %v", claims.Issuer, "test-issuer")
		}
	})

	t.Run("SetIssuer on MapClaims", func(t *testing.T) {
		claims := JWT.NewMapClaims()
		JWT.SetIssuer(claims, "test-issuer")

		if claims["iss"] != "test-issuer" {
			t.Errorf("SetIssuer() = %v, want %v", claims["iss"], "test-issuer")
		}
	})
}

func TestSetSubject(t *testing.T) {
	t.Run("SetSubject on StandardClaims", func(t *testing.T) {
		claims := JWT.NewStandardClaims()
		JWT.SetSubject(claims, "test-subject")

		if claims.Subject != "test-subject" {
			t.Errorf("SetSubject() = %v, want %v", claims.Subject, "test-subject")
		}
	})

	t.Run("SetSubject on MapClaims", func(t *testing.T) {
		claims := JWT.NewMapClaims()
		JWT.SetSubject(claims, "test-subject")

		if claims["sub"] != "test-subject" {
			t.Errorf("SetSubject() = %v, want %v", claims["sub"], "test-subject")
		}
	})
}

func TestSetAudience(t *testing.T) {
	t.Run("SetAudience with single audience on StandardClaims", func(t *testing.T) {
		claims := JWT.NewStandardClaims()
		JWT.SetAudience(claims, "audience1")

		if len(claims.Audience) != 1 || claims.Audience[0] != "audience1" {
			t.Errorf("SetAudience() = %v, want [audience1]", claims.Audience)
		}
	})

	t.Run("SetAudience with multiple audiences on StandardClaims", func(t *testing.T) {
		claims := JWT.NewStandardClaims()
		JWT.SetAudience(claims, "audience1", "audience2", "audience3")

		if len(claims.Audience) != 3 {
			t.Errorf("SetAudience() length = %d, want 3", len(claims.Audience))
		}
	})

	t.Run("SetAudience with single audience on MapClaims", func(t *testing.T) {
		claims := JWT.NewMapClaims()
		JWT.SetAudience(claims, "audience1")

		if claims["aud"] != "audience1" {
			t.Errorf("SetAudience() = %v, want %v", claims["aud"], "audience1")
		}
	})

	t.Run("SetAudience with multiple audiences on MapClaims", func(t *testing.T) {
		claims := JWT.NewMapClaims()
		JWT.SetAudience(claims, "audience1", "audience2")

		aud, ok := claims["aud"].([]string)
		if !ok || len(aud) != 2 || aud[0] != "audience1" || aud[1] != "audience2" {
			t.Errorf("SetAudience() = %v, want [audience1, audience2]", claims["aud"])
		}
	})
}

func TestAddCustomClaim(t *testing.T) {
	t.Run("AddCustomClaim to MapClaims", func(t *testing.T) {
		claims := JWT.NewMapClaims()
		JWT.AddCustomClaim(claims, "custom_key", "custom_value")

		if claims["custom_key"] != "custom_value" {
			t.Errorf("AddCustomClaim() = %v, want %v", claims["custom_key"], "custom_value")
		}
	})

	t.Run("AddCustomClaim with various types", func(t *testing.T) {
		claims := JWT.NewMapClaims()

		JWT.AddCustomClaim(claims, "string", "hello")
		JWT.AddCustomClaim(claims, "number", 12345)
		JWT.AddCustomClaim(claims, "float", 123.45)
		JWT.AddCustomClaim(claims, "bool", true)
		JWT.AddCustomClaim(claims, "nil", nil)

		if claims["string"] != "hello" {
			t.Errorf("string claim = %v, want hello", claims["string"])
		}
		if claims["number"] != 12345 {
			t.Errorf("number claim = %v, want 12345", claims["number"])
		}
		if claims["float"] != 123.45 {
			t.Errorf("float claim = %v, want 123.45", claims["float"])
		}
		if claims["bool"] != true {
			t.Errorf("bool claim = %v, want true", claims["bool"])
		}
		if claims["nil"] != nil {
			t.Errorf("nil claim = %v, want nil", claims["nil"])
		}
	})

	t.Run("AddCustomClaim overwrites existing", func(t *testing.T) {
		claims := JWT.NewMapClaims()
		claims["key"] = "old_value"

		JWT.AddCustomClaim(claims, "key", "new_value")

		if claims["key"] != "new_value" {
			t.Errorf("AddCustomClaim() = %v, want %v", claims["key"], "new_value")
		}
	})
}

// =========================================
// Test: Integration Tests
// =========================================

func TestJWTIntegration_CompleteFlow(t *testing.T) {
	secretKey := []byte("integration-test-secret")

	// Create claims using convenience methods
	claims := JWT.NewStandardClaims()
	JWT.SetIssuer(claims, "integration-issuer")
	JWT.SetSubject(claims, "integration-subject")
	JWT.SetAudience(claims, "integration-audience")
	JWT.SetExpiration(claims, time.Hour)
	JWT.SetIssuedAt(claims, time.Now())

	// Generate token
	token, err := JWT.GenerateHS256Token(claims, secretKey)
	if err != nil {
		t.Fatalf("GenerateHS256Token() error = %v", err)
	}

	// Parse token
	parsedClaims, err := JWT.ParseHS256Token(token, secretKey)
	if err != nil {
		t.Fatalf("ParseHS256Token() error = %v", err)
	}

	// Validate token
	err = JWT.ValidateClaims(parsedClaims,
		WithIssuer("integration-issuer"),
		WithSubject("integration-subject"),
		WithAudience("integration-audience"),
	)
	if err != nil {
		t.Fatalf("ValidateClaims() error = %v", err)
	}

	// Verify all fields
	if parsedClaims.GetIssuer() != "integration-issuer" {
		t.Errorf("Issuer = %v, want integration-issuer", parsedClaims.GetIssuer())
	}
	if parsedClaims.GetSubject() != "integration-subject" {
		t.Errorf("Subject = %v, want integration-subject", parsedClaims.GetSubject())
	}
	if len(parsedClaims.GetAudience()) != 1 || parsedClaims.GetAudience()[0] != "integration-audience" {
		t.Errorf("Audience = %v, want [integration-audience]", parsedClaims.GetAudience())
	}
}

func TestJWTIntegration_MapClaimsWithCustomFields(t *testing.T) {
	secretKey := []byte("integration-test-secret")

	// Create claims with custom fields
	claims := JWT.NewMapClaims()
	JWT.SetIssuer(claims, "custom-issuer")
	JWT.SetSubject(claims, "user123")
	JWT.SetExpiration(claims, 24*time.Hour)
	JWT.AddCustomClaim(claims, "user_id", 12345)
	JWT.AddCustomClaim(claims, "role", "admin")
	JWT.AddCustomClaim(claims, "permissions", []string{"read", "write", "delete"})

	// Generate token
	token, err := JWT.GenerateHS256Token(claims, secretKey)
	if err != nil {
		t.Fatalf("GenerateHS256Token() error = %v", err)
	}

	// Parse token
	parsedClaims, err := JWT.ParseHS256Token(token, secretKey)
	if err != nil {
		t.Fatalf("ParseHS256Token() error = %v", err)
	}

	// Verify custom fields
	customClaims := parsedClaims.GetCustomClaims()
	if len(customClaims) != 3 {
		t.Errorf("GetCustomClaims() returned %d fields, want 3", len(customClaims))
	}
	// JSON unmarshaling converts numbers to float64
	if customClaims["user_id"] != float64(12345) {
		t.Errorf("user_id = %v (type %T), want 12345", customClaims["user_id"], customClaims["user_id"])
	}
	if customClaims["role"] != "admin" {
		t.Errorf("role = %v, want admin", customClaims["role"])
	}
}

// =========================================
// Test: Edge Cases
// =========================================

func TestJWT_EdgeCase_TokenWithoutExpiration(t *testing.T) {
	secretKey := []byte("test-secret")

	claims := MapClaims{
		"iss": "test-issuer",
		"sub": "test-subject",
	}

	token, err := JWT.GenerateHS256Token(claims, secretKey)
	if err != nil {
		t.Fatalf("GenerateHS256Token() error = %v", err)
	}

	parsedClaims, err := JWT.ParseHS256Token(token, secretKey)
	if err != nil {
		t.Fatalf("ParseHS256Token() error = %v", err)
	}

	// Should not error - token without expiration is valid
	err = JWT.ValidateClaims(parsedClaims)
	if err != nil {
		t.Errorf("ValidateClaims() should not error for token without expiration, got %v", err)
	}
}

func TestJWT_EdgeCase_VeryLargeExpiration(t *testing.T) {
	secretKey := []byte("test-secret")

	claims := MapClaims{
		"exp": float64(time.Now().Add(365 * 24 * time.Hour).Unix()), // 1 year
	}

	token, err := JWT.GenerateHS256Token(claims, secretKey)
	if err != nil {
		t.Fatalf("GenerateHS256Token() error = %v", err)
	}

	parsedClaims, err := JWT.ParseHS256Token(token, secretKey)
	if err != nil {
		t.Fatalf("ParseHS256Token() error = %v", err)
	}

	err = JWT.ValidateClaims(parsedClaims)
	if err != nil {
		t.Errorf("ValidateClaims() should not error for token with far future expiration, got %v", err)
	}
}

func TestJWT_EdgeCase_UnicodeClaims(t *testing.T) {
	secretKey := []byte("test-secret")

	claims := MapClaims{
		"iss":     "测试签发者",
		"sub":     "测试主题",
		"message": "Hello 世界 🌍",
	}

	token, err := JWT.GenerateHS256Token(claims, secretKey)
	if err != nil {
		t.Fatalf("GenerateHS256Token() error = %v", err)
	}

	parsedClaims, err := JWT.ParseHS256Token(token, secretKey)
	if err != nil {
		t.Fatalf("ParseHS256Token() error = %v", err)
	}

	if parsedClaims.GetIssuer() != "测试签发者" {
		t.Errorf("Issuer = %v, want 测试签发者", parsedClaims.GetIssuer())
	}
	if parsedClaims.GetSubject() != "测试主题" {
		t.Errorf("Subject = %v, want 测试主题", parsedClaims.GetSubject())
	}
	if parsedClaims["message"] != "Hello 世界 🌍" {
		t.Errorf("message = %v, want Hello 世界 🌍", parsedClaims["message"])
	}
}

func TestJWT_EdgeCase_EmptyClaims(t *testing.T) {
	secretKey := []byte("test-secret")

	claims := MapClaims{}

	token, err := JWT.GenerateHS256Token(claims, secretKey)
	if err != nil {
		t.Fatalf("GenerateHS256Token() error = %v", err)
	}

	parsedClaims, err := JWT.ParseHS256Token(token, secretKey)
	if err != nil {
		t.Fatalf("ParseHS256Token() error = %v", err)
	}

	if len(parsedClaims) != 0 {
		t.Errorf("Parsed claims should be empty, got %d fields", len(parsedClaims))
	}
}

// =========================================
// Test: Algorithm Constants
// =========================================

func TestJWTAlgorithm_Constants(t *testing.T) {
	tests := []struct {
		name  string
		value JWTAlgorithm
		want  string
	}{
		{"HS256", HS256, "HS256"},
		{"HS384", HS384, "HS384"},
		{"HS512", HS512, "HS512"},
		{"RS256", RS256, "RS256"},
		{"RS384", RS384, "RS384"},
		{"RS512", RS512, "RS512"},
		{"ES256", ES256, "ES256"},
		{"ES384", ES384, "ES384"},
		{"ES512", ES512, "ES512"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.value, tt.want)
			}
		})
	}
}

// =========================================
// Benchmark Tests
// =========================================

func BenchmarkGenerateHS256Token(b *testing.B) {
	secretKey := []byte("benchmark-secret-key")
	claims := MapClaims{
		"iss": "benchmark-issuer",
		"sub": "benchmark-subject",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = JWT.GenerateHS256Token(claims, secretKey)
	}
}

func BenchmarkParseHS256Token(b *testing.B) {
	secretKey := []byte("benchmark-secret-key")
	claims := MapClaims{
		"iss": "benchmark-issuer",
		"sub": "benchmark-subject",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}
	token, _ := JWT.GenerateHS256Token(claims, secretKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = JWT.ParseHS256Token(token, secretKey)
	}
}

func BenchmarkGenerateRS256Token(b *testing.B) {
	privateKey, _, err := generateTestRSAKey()
	if err != nil {
		b.Fatalf("Failed to generate RSA key: %v", err)
	}

	claims := MapClaims{
		"iss": "benchmark-issuer",
		"sub": "benchmark-subject",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = JWT.GenerateRS256Token(claims, privateKey)
	}
}

func BenchmarkParseRS256Token(b *testing.B) {
	privateKey, publicKey, err := generateTestRSAKey()
	if err != nil {
		b.Fatalf("Failed to generate RSA key: %v", err)
	}

	claims := MapClaims{
		"iss": "benchmark-issuer",
		"sub": "benchmark-subject",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}
	token, _ := JWT.GenerateRS256Token(claims, privateKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = JWT.ParseRS256Token(token, publicKey)
	}
}

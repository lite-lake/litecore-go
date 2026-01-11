package observability

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"

	"com.litelake.litecore/manager/loggermgr"
)

// observabilityPlugin GORM 可观测性插件
type observabilityPlugin struct {
	logger               loggermgr.Logger
	tracer               trace.Tracer
	meter                metric.Meter
	queryDuration        metric.Float64Histogram
	queryCount           metric.Int64Counter
	queryErrorCount      metric.Int64Counter
	slowQueryCount       metric.Int64Counter
	transactionCount     metric.Int64Counter
	connectionPool       metric.Float64Gauge
	slowQueryThreshold   time.Duration
	logSQL               bool
	sampleRate           float64
}

// Setup 设置观测组件和指标
func (p *observabilityPlugin) Setup(
	logger loggermgr.Logger,
	tracer trace.Tracer,
	meter metric.Meter,
	queryDuration metric.Float64Histogram,
	queryCount metric.Int64Counter,
	queryErrorCount metric.Int64Counter,
	slowQueryCount metric.Int64Counter,
	transactionCount metric.Int64Counter,
	connectionPool metric.Float64Gauge,
) {
	p.logger = logger
	p.tracer = tracer
	p.meter = meter
	p.queryDuration = queryDuration
	p.queryCount = queryCount
	p.queryErrorCount = queryErrorCount
	p.slowQueryCount = slowQueryCount
	p.transactionCount = transactionCount
	p.connectionPool = connectionPool
}

// SetConfig 设置可观测性配置
func (p *observabilityPlugin) SetConfig(slowQueryThreshold time.Duration, logSQL bool, sampleRate float64) {
	p.slowQueryThreshold = slowQueryThreshold
	p.logSQL = logSQL
	p.sampleRate = sampleRate
}

// Name 插件名称
func (p *observabilityPlugin) Name() string {
	return "observability"
}

// Initialize GORM 插件初始化
func (p *observabilityPlugin) Initialize(db *gorm.DB) error {
	// 注册 Callback
	if p.tracer != nil || p.logger != nil {
		p.registerCallbacks(db)
	}
	return nil
}

// registerCallbacks 注册回调
func (p *observabilityPlugin) registerCallbacks(db *gorm.DB) {
	// 查询回调
	db.Callback().Query().Before("gorm:query").Register("observability:before_query", p.beforeQuery)
	db.Callback().Query().After("gorm:query").Register("observability:after_query", p.afterQuery)

	// 创建回调
	db.Callback().Create().Before("gorm:create").Register("observability:before_create", p.beforeCreate)
	db.Callback().Create().After("gorm:create").Register("observability:after_create", p.afterCreate)

	// 更新回调
	db.Callback().Update().Before("gorm:update").Register("observability:before_update", p.beforeUpdate)
	db.Callback().Update().After("gorm:update").Register("observability:after_update", p.afterUpdate)

	// 删除回调
	db.Callback().Delete().Before("gorm:delete").Register("observability:before_delete", p.beforeDelete)
	db.Callback().Delete().After("gorm:delete").Register("observability:after_delete", p.afterDelete)

	// 注意：GORM 的事务操作没有回调接口，事务追踪需要在应用层处理
	// 或通过 GORM 插件的其他方式实现
}

// Query 操作
func (p *observabilityPlugin) beforeQuery(db *gorm.DB) {
	p.recordOperationStart(db, "query")
}

func (p *observabilityPlugin) afterQuery(db *gorm.DB) {
	p.recordOperationEnd(db, "query", db.Error)
}

// Create 操作
func (p *observabilityPlugin) beforeCreate(db *gorm.DB) {
	p.recordOperationStart(db, "create")
}

func (p *observabilityPlugin) afterCreate(db *gorm.DB) {
	p.recordOperationEnd(db, "create", db.Error)
}

// Update 操作
func (p *observabilityPlugin) beforeUpdate(db *gorm.DB) {
	p.recordOperationStart(db, "update")
}

func (p *observabilityPlugin) afterUpdate(db *gorm.DB) {
	p.recordOperationEnd(db, "update", db.Error)
}

// Delete 操作
func (p *observabilityPlugin) beforeDelete(db *gorm.DB) {
	p.recordOperationStart(db, "delete")
}

func (p *observabilityPlugin) afterDelete(db *gorm.DB) {
	p.recordOperationEnd(db, "delete", db.Error)
}

// recordOperationStart 记录操作开始
func (p *observabilityPlugin) recordOperationStart(db *gorm.DB, operation string) {
	// 如果没有观测组件，直接返回
	if p.tracer == nil && p.logger == nil {
		return
	}

	// 采样检查
	if p.sampleRate < 1.0 && rand.Float64() > p.sampleRate {
		return
	}

	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	var span trace.Span
	if p.tracer != nil {
		ctx, span = p.tracer.Start(ctx, "db."+operation,
			trace.WithAttributes(
				attribute.String("db.operation", operation),
				attribute.String("db.table", db.Statement.Table),
			),
		)
		// 将新 context 设置回 db.Statement
		db.Statement.Context = ctx
	}

	// 记录开始时间
	db.InstanceSet("observability:start_time", time.Now())
	db.InstanceSet("observability:span", span)
}

// recordOperationEnd 记录操作结束
func (p *observabilityPlugin) recordOperationEnd(db *gorm.DB, operation string, err error) {
	// 获取开始时间
	startTime, ok := db.InstanceGet("observability:start_time")
	if !ok {
		return
	}

	start, _ := startTime.(time.Time)
	duration := time.Since(start).Seconds()

	// 获取 span
	spanInterface, _ := db.InstanceGet("observability:span")
	var span trace.Span
	if spanInterface != nil {
		span = spanInterface.(trace.Span)
	}

	// 记录指标
	if p.meter != nil {
		attrs := []attribute.KeyValue{
			attribute.String("operation", operation),
			attribute.String("table", db.Statement.Table),
			attribute.String("status", getStatus(err)),
		}

		// 记录查询耗时
		if p.queryDuration != nil {
			p.queryDuration.Record(db.Statement.Context, duration, metric.WithAttributes(attrs...))
		}

		// 记录查询计数
		if p.queryCount != nil {
			p.queryCount.Add(db.Statement.Context, 1, metric.WithAttributes(attrs...))
		}

		// 记录错误
		if err != nil && p.queryErrorCount != nil {
			p.queryErrorCount.Add(db.Statement.Context, 1, metric.WithAttributes(attrs...))
		}

		// 记录慢查询
		if p.slowQueryCount != nil && p.slowQueryThreshold > 0 {
			if time.Since(start) >= p.slowQueryThreshold {
				p.slowQueryCount.Add(db.Statement.Context, 1, metric.WithAttributes(attrs...))
			}
		}

		// 记录事务计数
		if p.transactionCount != nil && (operation == "commit" || operation == "rollback") {
			p.transactionCount.Add(db.Statement.Context, 1, metric.WithAttributes(attrs...))
		}
	}

	// 记录日志
	if p.logger != nil {
		if err != nil {
			logArgs := []any{
				"operation", operation,
				"table", db.Statement.Table,
				"error", err.Error(),
				"duration", duration,
			}
			if p.logSQL {
				logArgs = append(logArgs, "sql", sanitizeSQL(db.Statement.SQL.String()))
			}
			p.logger.Error("database operation failed", logArgs...)
			if span != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
		} else {
			// 慢查询使用 Warn 级别
			if p.slowQueryThreshold > 0 && time.Since(start) >= p.slowQueryThreshold {
				logArgs := []any{
					"operation", operation,
					"table", db.Statement.Table,
					"duration", duration,
					"threshold", p.slowQueryThreshold.Seconds(),
				}
				if p.logSQL {
					logArgs = append(logArgs, "sql", sanitizeSQL(db.Statement.SQL.String()))
				}
				p.logger.Warn("slow database query detected", logArgs...)
			} else {
				// 正常操作使用 Debug 级别
				p.logger.Debug("database operation success",
					"operation", operation,
					"table", db.Statement.Table,
					"duration", duration,
				)
			}
		}
	}

	// 结束 span
	if span != nil {
		span.End()
	}
}

// getStatus 根据错误获取状态
func getStatus(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

// sanitizeSQL 脱敏 SQL 语句中的敏感信息
func sanitizeSQL(sql string) string {
	if sql == "" {
		return ""
	}

	// 限制 SQL 语句长度（避免日志过大）
	const maxSQLLength = 500
	if len(sql) > maxSQLLength {
		sql = sql[:maxSQLLength] + "..."
	}

	// 脱敏密码参数（常见模式）
	passwordPatterns := []string{
		`password\s*=\s*'[^']*'`,
		`password\s*=\s*"[^"]*"`,
		`pwd\s*=\s*'[^']*'`,
		`pwd\s*=\s*"[^"]*"`,
		`token\s*=\s*'[^']*'`,
		`token\s*=\s*"[^"]*"`,
		`secret\s*=\s*'[^']*'`,
		`secret\s*=\s*"[^"]*"`,
		`api_key\s*=\s*'[^']*'`,
		`api_key\s*=\s*"[^"]*"`,
	}

	for _, pattern := range passwordPatterns {
		re := regexp.MustCompile(`(?i)`+pattern)
		sql = re.ReplaceAllString(sql, "***")
	}

	// 脱敏字符串值中的敏感字段（简单版本）
	// 更完整的版本需要解析 SQL AST
	sensitiveFields := []string{"password", "pwd", "token", "secret", "api_key"}
	for _, field := range sensitiveFields {
		// 匹配 field = 'value' 或 field = "value"
		re := regexp.MustCompile(`(?i)(`+field+`\s*=\s*)'[^']*'`)
		sql = re.ReplaceAllString(sql, "$1'***'")
		re = regexp.MustCompile(`(?i)(`+field+`\s*=\s*)"[^"]*"`)
		sql = re.ReplaceAllString(sql, `$1"***"`)
	}

	return strings.TrimSpace(sql)
}

// GetObservabilityPlugin 导出插件构造函数（供 manager.go 使用）
func NewObservabilityPlugin() *observabilityPlugin {
	return &observabilityPlugin{
		slowQueryThreshold: 1 * time.Second,
		logSQL:             false,
		sampleRate:         1.0,
	}
}

// Helper function to create common attributes
func getCommonAttributes(operation, table string, err error) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.String("table", table),
		attribute.String("status", getStatus(err)),
	}
}

// GetSlowQueryKey 获取慢查询的唯一标识
func GetSlowQueryKey(table, operation string) string {
	return fmt.Sprintf("%s:%s", table, operation)
}

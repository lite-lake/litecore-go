package databasemgr

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lite-lake/litecore-go/common"
	gormlogger "gorm.io/gorm/logger"
)

// ========== 测试实体 ==========

// gateTestUser 指纹门控测试实体
type gateTestUser struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"size:255;not null"`
	Email string `gorm:"size:255;uniqueIndex"`
}

func (gateTestUser) TableName() string { return "gate_test_users" }

// gateTestOrder 带外键索引的测试实体
type gateTestOrder struct {
	ID     uint `gorm:"primarykey"`
	UserID uint `gorm:"index"`
	Amount float64
}

func (gateTestOrder) TableName() string { return "gate_test_orders" }

// gateModelV1 / gateModelV2 同一表不同结构的变更测试实体
type gateModelV1 struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"size:100"`
}

func (gateModelV1) TableName() string { return "gate_models" }

type gateModelV2 struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"size:100"`
	Extra string
}

func (gateModelV2) TableName() string { return "gate_models" }

// ========== 指纹计算单元测试 ==========

// TestComputeSchemaFingerprint_Stable 指纹确定性：同一 struct 多次计算 hash 一致
func TestComputeSchemaFingerprint_Stable(t *testing.T) {
	table1, fp1, err := computeSchemaFingerprint(&gateTestUser{})
	if err != nil {
		t.Fatalf("computeSchemaFingerprint() 首次计算失败: %v", err)
	}

	table2, fp2, err := computeSchemaFingerprint(&gateTestUser{})
	if err != nil {
		t.Fatalf("computeSchemaFingerprint() 再次计算失败: %v", err)
	}

	if table1 != table2 {
		t.Errorf("表名不一致: %s != %s", table1, table2)
	}
	if fp1 != fp2 {
		t.Errorf("指纹不稳定: %s != %s", fp1, fp2)
	}
	if len(fp1) != 64 {
		t.Errorf("指纹长度应为 64（SHA256 hex），实际 %d", len(fp1))
	}
}

// TestComputeSchemaFingerprint_FieldOrderInsensitive 指纹确定性：字段声明顺序不影响 hash
func TestComputeSchemaFingerprint_FieldOrderInsensitive(t *testing.T) {
	_, fpA, err := computeSchemaFingerprint(&fingerprintOrderA{})
	if err != nil {
		t.Fatalf("computeSchemaFingerprint(orderA) 失败: %v", err)
	}
	_, fpB, err := computeSchemaFingerprint(&fingerprintOrderB{})
	if err != nil {
		t.Fatalf("computeSchemaFingerprint(orderB) 失败: %v", err)
	}

	if fpA != fpB {
		t.Errorf("字段声明顺序不同不应改变指纹:\n  A=%s\n  B=%s", fpA, fpB)
	}
}

// fingerprintOrderA / fingerprintOrderB 字段声明顺序不同但结构相同的实体
type fingerprintOrderA struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"size:100"`
	Age  int
}

func (fingerprintOrderA) TableName() string { return "fingerprint_models" }

type fingerprintOrderB struct {
	Age  int
	Name string `gorm:"size:100"`
	ID   uint   `gorm:"primarykey"`
}

func (fingerprintOrderB) TableName() string { return "fingerprint_models" }

// TestComputeSchemaFingerprint_Sensitive 指纹敏感性：改列类型/加列/改注释/改索引 → hash 变化
func TestComputeSchemaFingerprint_Sensitive(t *testing.T) {
	baseTable, baseFP, err := computeSchemaFingerprint(&fingerprintSensitiveBase{})
	if err != nil {
		t.Fatalf("computeSchemaFingerprint(base) 失败: %v", err)
	}

	cases := []struct {
		name string
		mod  any
	}{
		{"改列类型", &fingerprintSensitiveType{}},
		{"增加列", &fingerprintSensitiveColumn{}},
		{"修改注释", &fingerprintSensitiveComment{}},
		{"修改索引", &fingerprintSensitiveIndex{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			table, fp, err := computeSchemaFingerprint(tc.mod)
			if err != nil {
				t.Fatalf("computeSchemaFingerprint() 失败: %v", err)
			}
			if table != baseTable {
				t.Errorf("表名应相同: %s != %s", table, baseTable)
			}
			if fp == baseFP {
				t.Errorf("%s 不应改变指纹", tc.name)
			}
		})
	}
}

// 敏感性测试实体（同一张表不同结构）
type fingerprintSensitiveBase struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"size:255;comment:用户姓名"`
}

func (fingerprintSensitiveBase) TableName() string { return "fingerprint_sensitive" }

type fingerprintSensitiveType struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"size:100;comment:用户姓名"`
}

func (fingerprintSensitiveType) TableName() string { return "fingerprint_sensitive" }

type fingerprintSensitiveColumn struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"size:255;comment:用户姓名"`
	Email string `gorm:"size:64"`
}

func (fingerprintSensitiveColumn) TableName() string { return "fingerprint_sensitive" }

type fingerprintSensitiveComment struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"size:255;comment:用户昵称"`
}

func (fingerprintSensitiveComment) TableName() string { return "fingerprint_sensitive" }

type fingerprintSensitiveIndex struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"size:255;comment:用户姓名;uniqueIndex"`
}

func (fingerprintSensitiveIndex) TableName() string { return "fingerprint_sensitive" }

// gateConfigEntity 嵌入框架基类的测试实体（模拟真实应用实体）
type gateConfigEntity struct {
	common.BaseEntityWithTimestamps
	Key   string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Value string `gorm:"type:text;not null"`
}

func (gateConfigEntity) TableName() string { return "gate_configs" }

// TestComputeSchemaFingerprint_BaseEntity 嵌入框架基类实体（BaseEntityWithTimestamps）的指纹稳定且可迁移
func TestComputeSchemaFingerprint_BaseEntity(t *testing.T) {
	_, fp1, err := computeSchemaFingerprint(&gateConfigEntity{})
	if err != nil {
		t.Fatalf("computeSchemaFingerprint() error = %v", err)
	}
	_, fp2, err := computeSchemaFingerprint(&gateConfigEntity{})
	if err != nil {
		t.Fatalf("computeSchemaFingerprint() error = %v", err)
	}
	if fp1 != fp2 {
		t.Errorf("嵌入基类的实体指纹应稳定: %s != %s", fp1, fp2)
	}
}

// TestMigrateWithGate_BaseEntity 嵌入框架基类实体的门控迁移端到端验证
func TestMigrateWithGate_BaseEntity(t *testing.T) {
	skipIfCGONotAvailable(t)
	mgr := newGateSQLiteManager(t)

	counter := &sqlCounter{base: mgr.DB().Logger}
	mgr.DB().Logger = counter

	if err := mgr.AutoMigrate(&gateConfigEntity{}); err != nil {
		t.Fatalf("首次 AutoMigrate() error = %v", err)
	}

	counter.reset()
	if err := mgr.AutoMigrate(&gateConfigEntity{}); err != nil {
		t.Fatalf("第二次 AutoMigrate() error = %v", err)
	}
	if c := counter.snapshot(); c > 2 {
		t.Errorf("第二次应命中门控（<=2 SQL），实际 %d", c)
	}
}

// TestFilterChanged 变更集合过滤
func TestFilterChanged(t *testing.T) {
	computed := []pending{
		{table: "users", fingerprint: "fp-a"},
		{table: "orders", fingerprint: "fp-b"},
	}

	t.Run("无变更", func(t *testing.T) {
		changed := filterChanged(computed, map[string]string{
			"users":  "fp-a",
			"orders": "fp-b",
		})
		if len(changed) != 0 {
			t.Errorf("应无变更，实际 %d", len(changed))
		}
	})

	t.Run("部分变更", func(t *testing.T) {
		changed := filterChanged(computed, map[string]string{
			"users": "fp-a",
		})
		if len(changed) != 1 || changed[0].table != "orders" {
			t.Errorf("应仅 orders 变更，实际 %+v", changed)
		}
	})

	t.Run("未记录视为变更", func(t *testing.T) {
		changed := filterChanged(computed, map[string]string{})
		if len(changed) != 2 {
			t.Errorf("未记录的表应视为变更，实际 %d", len(changed))
		}
	})
}

// ========== SQL 计数 logger ==========

// sqlCounter 统计 GORM 执行的 SQL 条数
type sqlCounter struct {
	base  gormlogger.Interface
	count int64
}

func (c *sqlCounter) LogMode(level gormlogger.LogLevel) gormlogger.Interface { return c }

func (c *sqlCounter) Info(ctx context.Context, msg string, data ...interface{}) {
	if c.base != nil {
		c.base.Info(ctx, msg, data...)
	}
}

func (c *sqlCounter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if c.base != nil {
		c.base.Warn(ctx, msg, data...)
	}
}

func (c *sqlCounter) Error(ctx context.Context, msg string, data ...interface{}) {
	if c.base != nil {
		c.base.Error(ctx, msg, data...)
	}
}

func (c *sqlCounter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	atomic.AddInt64(&c.count, 1)
	if c.base != nil {
		c.base.Trace(ctx, begin, fc, err)
	}
}

func (c *sqlCounter) snapshot() int64 { return atomic.LoadInt64(&c.count) }
func (c *sqlCounter) reset()          { atomic.StoreInt64(&c.count, 0) }

// ========== 门控集成测试（SQLite 内存库） ==========

// newGateSQLiteManager 创建带单连接池的 SQLite 内存库管理器（保证 :memory: 唯一）
func newGateSQLiteManager(t *testing.T) IDatabaseManager {
	t.Helper()
	mgr, err := NewDatabaseManagerSQLiteImpl(&SQLiteConfig{
		DSN: ":memory:",
		PoolConfig: &PoolConfig{
			MaxOpenConns:    1,
			MaxIdleConns:    1,
			ConnMaxLifetime: time.Minute,
			ConnMaxIdleTime: time.Minute,
		},
	}, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerSQLiteImpl() error = %v", err)
	}
	t.Cleanup(func() { _ = mgr.Close() })
	return mgr
}

// TestMigrateWithGate_SkipOnSecondCall 门控跳过：首次迁移后第二次调用 SQL 数大幅减少
func TestMigrateWithGate_SkipOnSecondCall(t *testing.T) {
	skipIfCGONotAvailable(t)
	mgr := newGateSQLiteManager(t)

	counter := &sqlCounter{base: mgr.DB().Logger}
	mgr.DB().Logger = counter

	models := []any{&gateTestUser{}, &gateTestOrder{}}

	if err := mgr.AutoMigrate(models...); err != nil {
		t.Fatalf("首次 AutoMigrate() error = %v", err)
	}
	firstCount := counter.snapshot()
	if firstCount < 4 {
		t.Errorf("首次迁移应执行较多 SQL（>=4），实际 %d", firstCount)
	}

	counter.reset()
	if err := mgr.AutoMigrate(models...); err != nil {
		t.Fatalf("第二次 AutoMigrate() error = %v", err)
	}
	secondCount := counter.snapshot()
	// 门控命中路径：仅 CREATE TABLE IF NOT EXISTS（1）+ SELECT 指纹（1）= 2 次 SQL
	if secondCount > 2 {
		t.Errorf("第二次迁移应命中门控（<=2 SQL），实际 %d", secondCount)
	}
}

// TestMigrateWithGate_RepeatedCallsConverge 连续多次调用最终收敛为门控命中
func TestMigrateWithGate_RepeatedCallsConverge(t *testing.T) {
	skipIfCGONotAvailable(t)
	mgr := newGateSQLiteManager(t)

	counter := &sqlCounter{base: mgr.DB().Logger}
	mgr.DB().Logger = counter

	models := []any{&gateTestUser{}, &gateTestOrder{}}
	for i := 0; i < 3; i++ {
		if err := mgr.AutoMigrate(models...); err != nil {
			t.Fatalf("第 %d 次 AutoMigrate() error = %v", i+1, err)
		}
		counter.reset()
	}
	if c := counter.snapshot(); c > 2 {
		t.Errorf("收敛后应命中门控（<=2 SQL），实际 %d", c)
	}
}

// TestMigrateWithGate_ChangeTriggersMigration 变更触发：改 struct 后再次调用仅迁移该表
func TestMigrateWithGate_ChangeTriggersMigration(t *testing.T) {
	skipIfCGONotAvailable(t)
	mgr := newGateSQLiteManager(t)

	if err := mgr.AutoMigrate(&gateModelV1{}); err != nil {
		t.Fatalf("AutoMigrate(v1) error = %v", err)
	}

	// 验证 v1 结构无 Extra 列
	if hasColumn(t, mgr, "gate_models", "extra") {
		t.Fatal("v1 迁移后不应存在 extra 列")
	}

	// 变更结构：增加 Extra 列
	if err := mgr.AutoMigrate(&gateModelV2{}); err != nil {
		t.Fatalf("AutoMigrate(v2) error = %v", err)
	}

	// 验证列已添加
	if !hasColumn(t, mgr, "gate_models", "extra") {
		t.Error("v2 迁移后应存在 extra 列")
	}
}

// hasColumn 检查 SQLite 表中是否存在指定列
func hasColumn(t *testing.T, mgr IDatabaseManager, table, column string) bool {
	t.Helper()
	var names []string
	if err := mgr.DB().Raw("SELECT name FROM pragma_table_info(?)", table).Scan(&names).Error; err != nil {
		t.Fatalf("读取表结构失败: %v", err)
	}
	for _, n := range names {
		if n == column {
			return true
		}
	}
	return false
}

// TestMigrateWithGate_ForceMigrate force_migrate 逃生阀：跳过指纹比对全量迁移
func TestMigrateWithGate_ForceMigrate(t *testing.T) {
	skipIfCGONotAvailable(t)
	mgr := newGateSQLiteManager(t)

	if gate, ok := mgr.(migrateGateConfigurer); ok {
		gate.setMigrateGateConfig(true, 30*time.Second)
	}

	counter := &sqlCounter{base: mgr.DB().Logger}
	mgr.DB().Logger = counter

	if err := mgr.AutoMigrate(&gateTestUser{}); err != nil {
		t.Fatalf("首次 AutoMigrate() error = %v", err)
	}
	firstCount := counter.snapshot()

	counter.reset()
	if err := mgr.AutoMigrate(&gateTestUser{}); err != nil {
		t.Fatalf("第二次 AutoMigrate() error = %v", err)
	}
	secondCount := counter.snapshot()

	if secondCount < 4 {
		t.Errorf("force_migrate 下第二次应全量迁移（>=4 SQL），实际 %d", secondCount)
	}
	if secondCount < firstCount {
		t.Errorf("force_migrate 下两次迁移开销应相当，first=%d second=%d", firstCount, secondCount)
	}
}

// TestMigrateWithGate_DegradeWhenMetaUnavailable 降级：schema_meta 表不可用 → 回退总是迁移，不阻断启动
func TestMigrateWithGate_DegradeWhenMetaUnavailable(t *testing.T) {
	skipIfCGONotAvailable(t)
	mgr := newGateSQLiteManager(t)

	// 预置一个结构错误的 litecore_schema_meta 表：CREATE IF NOT EXISTS 成功但 SELECT/upsert 失败
	if err := mgr.DB().Exec("CREATE TABLE litecore_schema_meta (x INTEGER)").Error; err != nil {
		t.Fatalf("预置坏表失败: %v", err)
	}

	if err := mgr.AutoMigrate(&gateTestUser{}); err != nil {
		t.Fatalf("降级路径不应报错: %v", err)
	}

	var count int64
	if err := mgr.DB().Table("gate_test_users").Count(&count).Error; err != nil {
		t.Fatalf("业务表应已创建: %v", err)
	}
}

// TestMigrateWithGate_ForceWithBrokenMeta force_migrate + schema_meta 损坏：迁移成功且不阻断启动
func TestMigrateWithGate_ForceWithBrokenMeta(t *testing.T) {
	skipIfCGONotAvailable(t)
	mgr := newGateSQLiteManager(t)

	if gate, ok := mgr.(migrateGateConfigurer); ok {
		gate.setMigrateGateConfig(true, 30*time.Second)
	}

	// 预置结构错误的 schema_meta 表
	if err := mgr.DB().Exec("CREATE TABLE litecore_schema_meta (x INTEGER)").Error; err != nil {
		t.Fatalf("预置坏表失败: %v", err)
	}

	if err := mgr.AutoMigrate(&gateTestUser{}); err != nil {
		t.Fatalf("force_migrate + 损坏 schema_meta 不应阻断启动: %v", err)
	}

	var count int64
	if err := mgr.DB().Table("gate_test_users").Count(&count).Error; err != nil {
		t.Fatalf("业务表应已创建: %v", err)
	}
}

// TestMigrateWithGate_NoneDriver none 驱动返回错误
func TestMigrateWithGate_NoneDriver(t *testing.T) {
	mgr := NewDatabaseManagerNoneImpl(nil, nil)
	err := mgr.AutoMigrate(&gateTestUser{})
	if err == nil {
		t.Error("none 驱动 AutoMigrate 应返回错误")
	}
}

// ========== 配置解析 ==========

// TestParseDatabaseConfigFromMap_MigrateGate 迁移门控配置解析
func TestParseDatabaseConfigFromMap_MigrateGate(t *testing.T) {
	tests := []struct {
		name  string
		cfg   map[string]any
		check func(*DatabaseConfig) bool
	}{
		{
			name: "默认值",
			cfg:  map[string]any{"driver": "sqlite"},
			check: func(c *DatabaseConfig) bool {
				return !c.ForceMigrate && c.MigrateLockTimeout == DefaultMigrateLockTimeout
			},
		},
		{
			name: "force_migrate true",
			cfg: map[string]any{
				"driver":        "sqlite",
				"force_migrate": true,
			},
			check: func(c *DatabaseConfig) bool { return c.ForceMigrate },
		},
		{
			name: "migrate_lock_timeout 字符串",
			cfg: map[string]any{
				"driver":               "sqlite",
				"migrate_lock_timeout": "1m",
			},
			check: func(c *DatabaseConfig) bool {
				return c.MigrateLockTimeout == time.Minute
			},
		},
		{
			name: "migrate_lock_timeout 整数秒",
			cfg: map[string]any{
				"driver":               "sqlite",
				"migrate_lock_timeout": 60,
			},
			check: func(c *DatabaseConfig) bool {
				return c.MigrateLockTimeout == 60*time.Second
			},
		},
		{
			name: "migrate_lock_timeout 非法格式",
			cfg: map[string]any{
				"driver":               "sqlite",
				"migrate_lock_timeout": "invalid",
			},
			check: func(c *DatabaseConfig) bool { return false },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseDatabaseConfigFromMap(tt.cfg)
			if tt.name == "migrate_lock_timeout 非法格式" {
				if err == nil {
					t.Error("非法时长格式应返回错误")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDatabaseConfigFromMap() error = %v", err)
			}
			if !tt.check(cfg) {
				t.Error("配置解析检查失败")
			}
		})
	}
}

// ========== 迁移锁（带真实 DB 时运行） ==========

// TestMigrateWithGate_LockMySQL MySQL 迁移锁加锁释放
// 需要环境变量 MYSQL_TEST_DSN（如 root:password@tcp(localhost:3306)/test）
func TestMigrateWithGate_LockMySQL(t *testing.T) {
	dsn := getenvOrSkip(t, "MYSQL_TEST_DSN")
	mgr, err := NewDatabaseManagerMySQLImpl(&MySQLConfig{DSN: dsn}, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerMySQLImpl() error = %v", err)
	}
	defer mgr.Close()

	base := mgr.(*databaseManagerMysqlImpl).databaseManagerBaseImpl
	ctx := context.Background()

	release, err := base.acquireMySQLLock(ctx)
	if err != nil {
		t.Fatalf("acquireMySQLLock() error = %v", err)
	}

	// 重复加锁：同一连接可重入
	release2, err := base.acquireMySQLLock(ctx)
	if err != nil {
		t.Fatalf("acquireMySQLLock() 第二次 error = %v", err)
	}
	release2()
	release()
}

// TestMigrateWithGate_LockPostgreSQL PostgreSQL 迁移锁加锁释放
// 需要环境变量 PG_TEST_DSN（如 host=localhost user=postgres password=x dbname=test sslmode=disable）
func TestMigrateWithGate_LockPostgreSQL(t *testing.T) {
	dsn := getenvOrSkip(t, "PG_TEST_DSN")
	mgr, err := NewDatabaseManagerPostgreSQLImpl(&PostgreSQLConfig{DSN: dsn}, nil, nil)
	if err != nil {
		t.Fatalf("NewDatabaseManagerPostgreSQLImpl() error = %v", err)
	}
	defer mgr.Close()

	base := mgr.(*databaseManagerPostgresqlImpl).databaseManagerBaseImpl
	ctx := context.Background()

	release, err := base.acquirePostgreSQLLock(ctx)
	if err != nil {
		t.Fatalf("acquirePostgreSQLLock() error = %v", err)
	}
	release()
}

// getenvOrSkip 读取环境变量，未设置则跳过测试
func getenvOrSkip(t *testing.T, key string) string {
	t.Helper()
	if v := os.Getenv(key); v != "" {
		return v
	}
	t.Skipf("环境变量 %s 未设置，跳过测试", key)
	return ""
}

// TestMigrateWithGate_AutoMigrateWithConfigProvider 通过 BuildWithConfigProvider 注入门控配置
func TestMigrateWithGate_AutoMigrateWithConfigProvider(t *testing.T) {
	skipIfCGONotAvailable(t)

	provider := &MockConfigProvider{
		data: map[string]any{
			"database.driver": "sqlite",
			"database.sqlite_config": map[string]any{
				"dsn": ":memory:",
				"pool_config": map[string]any{
					"max_open_conns": 1,
					"max_idle_conns": 1,
				},
			},
			"database.force_migrate":        true,
			"database.migrate_lock_timeout": "1m",
		},
	}

	mgr, err := BuildWithConfigProvider(provider, nil, nil)
	if err != nil {
		t.Fatalf("BuildWithConfigProvider() error = %v", err)
	}
	defer mgr.Close()

	base := mgr.(*databaseManagerSqliteImpl).databaseManagerBaseImpl
	if !base.forceMigrate {
		t.Error("force_migrate 应从配置注入")
	}
	if base.migrateLockTimeout != time.Minute {
		t.Errorf("migrate_lock_timeout 应从配置注入，实际 %v", base.migrateLockTimeout)
	}
}

package databasemgr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/lite-lake/litecore-go/logger"
)

// schemaMetaTableName 迁移指纹记录表名（框架自建，不属于业务实体）
const schemaMetaTableName = "litecore_schema_meta"

// migrateLockName 迁移锁名称（MySQL GET_LOCK 的锁名）
const migrateLockName = "litecore_migrate_lock"

// migrateLockKey PostgreSQL advisory lock 的锁 key（"litecore" 的十六进制表示，正值，兼容 int8）
const migrateLockKey int64 = 0x6C697465636F7265

// schemaCache schema.Parse 的进程内缓存（按实体类型缓存，避免重复解析）
var schemaCache sync.Map

// ========== 指纹计算 ==========

// fieldFingerprint 字段指纹
type fieldFingerprint struct {
	DBName        string `json:"db_name"`
	DataType      string `json:"data_type"`
	GORMDataType  string `json:"gorm_data_type"`
	PrimaryKey    bool   `json:"primary_key"`
	NotNull       bool   `json:"not_null"`
	Unique        bool   `json:"unique"`
	AutoIncrement bool   `json:"auto_increment"`
	DefaultValue  string `json:"default_value"`
	Comment       string `json:"comment"`
	Size          int    `json:"size"`
	Precision     int    `json:"precision"`
	Scale         int    `json:"scale"`
}

// indexFieldFingerprint 索引字段指纹
type indexFieldFingerprint struct {
	DBName  string `json:"db_name"`
	Sort    string `json:"sort"`
	Collate string `json:"collate"`
	Length  int    `json:"length"`
}

// indexFingerprint 索引指纹
type indexFingerprint struct {
	Name   string                  `json:"name"`
	Class  string                  `json:"class"`
	Type   string                  `json:"type"`
	Where  string                  `json:"where"`
	Fields []indexFieldFingerprint `json:"fields"`
}

// checkConstraintFingerprint 检查约束指纹
type checkConstraintFingerprint struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint"`
}

// fkConstraintFingerprint 外键约束指纹
type fkConstraintFingerprint struct {
	Name        string   `json:"name"`
	ForeignKeys []string `json:"foreign_keys"`
	References  []string `json:"references"`
	OnDelete    string   `json:"on_delete"`
	OnUpdate    string   `json:"on_update"`
}

// schemaFingerprint 表结构指纹（确定性 JSON 序列化 → SHA256）
type schemaFingerprint struct {
	Table            string                       `json:"table"`
	Fields           []fieldFingerprint           `json:"fields"`
	Indexes          []indexFingerprint           `json:"indexes"`
	CheckConstraints []checkConstraintFingerprint `json:"check_constraints"`
	ForeignKeys      []fkConstraintFingerprint    `json:"foreign_keys"`
}

// computeSchemaFingerprint 计算实体表结构指纹（内存计算，不需要活连接）
// 返回：表名、指纹（SHA256 hex）、错误
// 序列化前按 DBName / 索引名 / 约束名排序，保证字段声明顺序不影响指纹
func computeSchemaFingerprint(model any) (string, string, error) {
	s, err := schema.Parse(model, &schemaCache, schema.NamingStrategy{})
	if err != nil {
		return "", "", fmt.Errorf("解析模型 schema 失败: %w", err)
	}

	fp := schemaFingerprint{
		Table: s.Table,
	}

	// 字段：跳过 IgnoreMigration 的虚拟字段，按 DBName 排序
	for _, f := range s.Fields {
		if f.DBName == "" || f.IgnoreMigration {
			continue
		}
		fp.Fields = append(fp.Fields, fieldFingerprint{
			DBName:        f.DBName,
			DataType:      string(f.DataType),
			GORMDataType:  string(f.GORMDataType),
			PrimaryKey:    f.PrimaryKey,
			NotNull:       f.NotNull,
			Unique:        f.Unique,
			AutoIncrement: f.AutoIncrement,
			DefaultValue:  f.DefaultValue,
			Comment:       f.Comment,
			Size:          f.Size,
			Precision:     f.Precision,
			Scale:         f.Scale,
		})
	}
	sort.Slice(fp.Fields, func(i, j int) bool {
		return fp.Fields[i].DBName < fp.Fields[j].DBName
	})

	// 索引：按索引名排序
	for _, idx := range s.ParseIndexes() {
		ifp := indexFingerprint{
			Name:  idx.Name,
			Class: idx.Class,
			Type:  idx.Type,
			Where: idx.Where,
		}
		for _, f := range idx.Fields {
			ifp.Fields = append(ifp.Fields, indexFieldFingerprint{
				DBName:  f.Field.DBName,
				Sort:    f.Sort,
				Collate: f.Collate,
				Length:  f.Length,
			})
		}
		fp.Indexes = append(fp.Indexes, ifp)
	}
	sort.Slice(fp.Indexes, func(i, j int) bool {
		return fp.Indexes[i].Name < fp.Indexes[j].Name
	})

	// 检查约束：按名称排序
	for _, chk := range s.ParseCheckConstraints() {
		fp.CheckConstraints = append(fp.CheckConstraints, checkConstraintFingerprint{
			Name:       chk.Name,
			Constraint: chk.Constraint,
		})
	}
	sort.Slice(fp.CheckConstraints, func(i, j int) bool {
		return fp.CheckConstraints[i].Name < fp.CheckConstraints[j].Name
	})

	// 外键约束：仅收录属于当前表的外键（与 GORM AutoMigrate 行为一致），按名称排序
	for _, rel := range s.Relationships.Relations {
		constraint := rel.ParseConstraint()
		if constraint == nil || constraint.Schema != s {
			continue
		}
		ffp := fkConstraintFingerprint{
			Name:     constraint.Name,
			OnDelete: constraint.OnDelete,
			OnUpdate: constraint.OnUpdate,
		}
		for _, fk := range constraint.ForeignKeys {
			ffp.ForeignKeys = append(ffp.ForeignKeys, fk.DBName)
		}
		for _, ref := range constraint.References {
			ffp.References = append(ffp.References, ref.DBName)
		}
		fp.ForeignKeys = append(fp.ForeignKeys, ffp)
	}
	sort.Slice(fp.ForeignKeys, func(i, j int) bool {
		return fp.ForeignKeys[i].Name < fp.ForeignKeys[j].Name
	})

	// 结构体字段声明顺序固定 + 数组已排序 → JSON 序列化确定性
	data, err := json.Marshal(fp)
	if err != nil {
		return "", "", fmt.Errorf("序列化 schema 指纹失败: %w", err)
	}

	sum := sha256.Sum256(data)
	return s.Table, hex.EncodeToString(sum[:]), nil
}

// ========== schema_meta 表读写（全部走原生 SQL，不建实体） ==========

// schemaMetaRow schema_meta 表行
type schemaMetaRow struct {
	TableName   string
	Fingerprint string
}

// schemaMetaDDL 返回三驱动兼容的 schema_meta 建表语句
func (b *databaseManagerBaseImpl) schemaMetaDDL() string {
	switch b.driver {
	case "postgresql":
		return `CREATE TABLE IF NOT EXISTS litecore_schema_meta (
    table_name  VARCHAR(128) NOT NULL,
    fingerprint CHAR(64)     NOT NULL,
    migrated_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (table_name)
)`
	case "sqlite":
		return `CREATE TABLE IF NOT EXISTS litecore_schema_meta (
    table_name  VARCHAR(128) NOT NULL,
    fingerprint CHAR(64)     NOT NULL,
    migrated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (table_name)
)`
	default: // mysql
		return `CREATE TABLE IF NOT EXISTS litecore_schema_meta (
    table_name  VARCHAR(128) NOT NULL,
    fingerprint CHAR(64)     NOT NULL,
    migrated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (table_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`
	}
}

// ensureSchemaMetaTable 懒创建 schema_meta 表；失败时返回错误（调用方降级为总是迁移）
func (b *databaseManagerBaseImpl) ensureSchemaMetaTable() error {
	if err := b.db.Exec(b.schemaMetaDDL()).Error; err != nil {
		return fmt.Errorf("创建迁移指纹表失败: %w", err)
	}
	return nil
}

// loadSchemaMeta 读取全部已记录指纹（table_name → fingerprint）
func (b *databaseManagerBaseImpl) loadSchemaMeta() (map[string]string, error) {
	var rows []schemaMetaRow
	if err := b.db.Table(schemaMetaTableName).Select("table_name", "fingerprint").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("读取迁移指纹失败: %w", err)
	}

	result := make(map[string]string, len(rows))
	for _, row := range rows {
		result[row.TableName] = row.Fingerprint
	}
	return result, nil
}

// upsertSchemaMeta 批量 upsert 指纹（单事务）
func (b *databaseManagerBaseImpl) upsertSchemaMeta(fingerprints map[string]string) error {
	if len(fingerprints) == 0 {
		return nil
	}

	var upsertSQL string
	switch b.driver {
	case "postgresql":
		upsertSQL = `INSERT INTO litecore_schema_meta (table_name, fingerprint) VALUES (?, ?)
			ON CONFLICT (table_name) DO UPDATE SET fingerprint = EXCLUDED.fingerprint, migrated_at = CURRENT_TIMESTAMP`
	case "sqlite":
		upsertSQL = `INSERT INTO litecore_schema_meta (table_name, fingerprint) VALUES (?, ?)
			ON CONFLICT (table_name) DO UPDATE SET fingerprint = excluded.fingerprint, migrated_at = CURRENT_TIMESTAMP`
	default: // mysql
		upsertSQL = `INSERT INTO litecore_schema_meta (table_name, fingerprint) VALUES (?, ?)
			ON DUPLICATE KEY UPDATE fingerprint = VALUES(fingerprint), migrated_at = CURRENT_TIMESTAMP`
	}

	return b.db.Transaction(func(tx *gorm.DB) error {
		for table, fp := range fingerprints {
			if err := tx.Exec(upsertSQL, table, fp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ========== 迁移锁（多副本串行化） ==========

// acquireMigrateLock 获取迁移锁，返回释放函数
// - MySQL：GET_LOCK / RELEASE_LOCK（独立连接）
// - PostgreSQL：pg_try_advisory_lock 轮询 / pg_advisory_unlock（独立连接）
// - SQLite：单写者，无需锁
func (b *databaseManagerBaseImpl) acquireMigrateLock(ctx context.Context) (func(), error) {
	switch b.driver {
	case "mysql":
		return b.acquireMySQLLock(ctx)
	case "postgresql":
		return b.acquirePostgreSQLLock(ctx)
	default:
		return func() {}, nil
	}
}

// acquireMySQLLock 获取 MySQL GET_LOCK
func (b *databaseManagerBaseImpl) acquireMySQLLock(ctx context.Context) (func(), error) {
	if b.sqlDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	conn, err := b.sqlDB.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取迁移锁连接失败: %w", err)
	}

	// GET_LOCK 超时参数单位为秒（支持小数），向上取整避免子秒配置被截断为 0
	timeout := int(math.Ceil(b.migrateLockTimeout.Seconds()))
	var acquired int
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", migrateLockName, timeout).Scan(&acquired); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("获取迁移锁失败: %w", err)
	}
	if acquired != 1 {
		_ = conn.Close()
		return nil, fmt.Errorf("获取迁移锁超时（%s），可能有其他实例正在迁移", b.migrateLockTimeout)
	}

	release := func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT RELEASE_LOCK(?)", migrateLockName)
		_ = conn.Close()
	}
	return release, nil
}

// acquirePostgreSQLLock 获取 PostgreSQL advisory lock
func (b *databaseManagerBaseImpl) acquirePostgreSQLLock(ctx context.Context) (func(), error) {
	if b.sqlDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	conn, err := b.sqlDB.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取迁移锁连接失败: %w", err)
	}

	deadline := time.Now().Add(b.migrateLockTimeout)
	for {
		var acquired bool
		if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(?)", migrateLockKey).Scan(&acquired); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("获取迁移锁失败: %w", err)
		}
		if acquired {
			release := func() {
				_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock(?)", migrateLockKey)
				_ = conn.Close()
			}
			return release, nil
		}

		if time.Now().After(deadline) {
			_ = conn.Close()
			return nil, fmt.Errorf("获取迁移锁超时（%s），可能有其他实例正在迁移", b.migrateLockTimeout)
		}

		select {
		case <-ctx.Done():
			_ = conn.Close()
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
}

// ========== 指纹门控主流程 ==========

// pending 待迁移模型及其指纹
type pending struct {
	model       any
	table       string
	fingerprint string
}

// filterChanged 过滤出指纹不匹配（或未记录）的模型
func filterChanged(computed []pending, meta map[string]string) []pending {
	changed := make([]pending, 0, len(computed))
	for _, p := range computed {
		if stored, ok := meta[p.table]; !ok || stored != p.fingerprint {
			changed = append(changed, p)
		}
	}
	return changed
}

// migrateWithGate 指纹门控迁移（所有 AutoMigrate 调用的必经点）
//
//  1. ensureSchemaMetaTable() 失败 → 降级为总是迁移 + WARN（不阻断启动）
//  2. 对每个 model 内存计算指纹（无数据库往返）
//  3. 一次 SELECT 读取已记录指纹，得到变更集合
//  4. 变更集合为空 → 毫秒级返回
//  5. 获取迁移锁，锁内二次校验（TOCTOU 防御）
//  6. AutoMigrate(changed...) 单次调用（保留 GORM 外键依赖排序）
//  7. 成功后 upsert 指纹，释放锁
func (b *databaseManagerBaseImpl) migrateWithGate(models ...any) error {
	if len(models) == 0 {
		return nil
	}
	if b.driver == "none" {
		return fmt.Errorf("database not available (none driver)")
	}

	b.migrateMu.Lock()
	defer b.migrateMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), b.migrateLockTimeout+5*time.Minute)
	defer cancel()

	// 1. 计算全部模型指纹（内存，不触发数据库往返）
	computed := make([]pending, 0, len(models))
	for _, m := range models {
		table, fp, err := computeSchemaFingerprint(m)
		if err != nil {
			return err
		}
		computed = append(computed, pending{model: m, table: table, fingerprint: fp})
	}

	// 2. 确保 schema_meta 表存在；失败 → 降级为总是迁移
	metaAvailable := true
	if err := b.ensureSchemaMetaTable(); err != nil {
		metaAvailable = false
		b.warn("迁移指纹表不可用，降级为总是迁移", "error", err)
	}

	// 逃生阀（force_migrate）或降级：跳过指纹比对，全量迁移
	skipGate := b.forceMigrate || !metaAvailable

	var changed []pending
	if skipGate {
		changed = computed
	} else {
		meta, err := b.loadSchemaMeta()
		if err != nil {
			metaAvailable = false
			b.warn("读取迁移指纹失败，降级为总是迁移", "error", err)
			changed = computed
			skipGate = true
		} else {
			changed = filterChanged(computed, meta)
			if len(changed) == 0 {
				b.info("迁移门控命中，无表结构变更，跳过 schema 迁移", "count", len(computed))
				return nil
			}
		}
	}

	// 3. 获取迁移锁（多副本串行化）
	release, err := b.acquireMigrateLock(ctx)
	if err != nil {
		return fmt.Errorf("获取迁移锁失败: %w", err)
	}
	defer release()

	// 4. 锁内二次校验（TOCTOU 防御）：仅门控路径需要
	if !skipGate {
		meta, err := b.loadSchemaMeta()
		if err != nil {
			// 与首次读取一致：降级为总是迁移，不阻断启动
			b.warn("锁内重读迁移指纹失败，降级为总是迁移", "error", err)
			changed = computed
		} else {
			changed = filterChanged(computed, meta)
			if len(changed) == 0 {
				b.info("迁移门控命中（锁内校验），无表结构变更，跳过 schema 迁移", "count", len(computed))
				return nil
			}
		}
	}

	// 5. 合并一次 AutoMigrate（保留 GORM 外键依赖排序）
	changedModels := make([]any, 0, len(changed))
	fingerprints := make(map[string]string, len(changed))
	for _, p := range changed {
		changedModels = append(changedModels, p.model)
		fingerprints[p.table] = p.fingerprint
	}

	b.info("开始数据库迁移", "count", len(changedModels))
	if err := b.db.AutoMigrate(changedModels...); err != nil {
		return fmt.Errorf("自动迁移失败: %w", err)
	}

	// 6. 记录指纹（schema_meta 可用时）；写失败不阻断启动（下次启动重复迁移，AutoMigrate 幂等）
	if metaAvailable {
		if err := b.upsertSchemaMeta(fingerprints); err != nil {
			b.warn("记录迁移指纹失败，下次启动将重新迁移", "error", err)
			return nil
		}
	}
	b.info("数据库迁移完成", "count", len(changedModels))

	return nil
}

// log 获取框架日志器（可能为 nil）
func (b *databaseManagerBaseImpl) log() logger.ILogger {
	if b.loggerMgr != nil {
		return b.loggerMgr.Ins()
	}
	return nil
}

// info 记录 Info 级别日志（loggerMgr 未注入时静默）
func (b *databaseManagerBaseImpl) info(msg string, args ...any) {
	if l := b.log(); l != nil {
		l.Info(msg, args...)
	}
}

// warn 记录 Warn 级别日志（loggerMgr 未注入时静默）
func (b *databaseManagerBaseImpl) warn(msg string, args ...any) {
	if l := b.log(); l != nil {
		l.Warn(msg, args...)
	}
}

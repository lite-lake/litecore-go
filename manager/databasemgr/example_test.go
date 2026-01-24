package databasemgr_test

import (
	"github.com/lite-lake/litecore-go/manager/databasemgr"
)

// Example_databasemgr_withFactory 演示如何使用工厂模式创建 IDatabaseManager
func Example_databasemgr_withFactory() {
	// 方式1: 直接使用 Build 函数
	cfg := map[string]any{
		"dsn": ":memory:",
	}
	dbMgr, _ := databasemgr.Build("sqlite", cfg, nil, nil)
	defer dbMgr.Close()

	// 使用数据库
	// db := dbMgr.DB()
	// var users []User
	// db.Find(&users)

	_ = dbMgr
}

// Example_databasemgr_withConfigProvider 演示如何使用 ConfigMgr 创建 IDatabaseManager
func Example_databasemgr_withConfigProvider() {
	// 方式2: 使用 ConfigMgr（推荐用于依赖注入场景）
	// import loggermgr "github.com/lite-lake/litecore-go/component/manager/loggermgr"
	// provider := configmgr.NewYamlConfigProvider("config.yaml")
	// dbMgr, err := databasemgr.BuildWithConfigProvider(provider)
	// if err != nil {
	//     loggerMgr := loggermgr.GetLoggerManager()
	//     loggerMgr.Logger("main").Fatal("Failed to create database manager", "error", err)
	// }
	// defer dbMgr.Close()

	// 使用数据库
	// db := dbMgr.DB()
	// var users []User
	// db.Find(&users)
}

// Example_databasemgr_configuration 演示配置格式
func Example_databasemgr_configuration() {
	// 配置文件示例（YAML 格式）：
	//
	// database:
	//   driver: mysql
	//   mysql_config:
	//     dsn: "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
	//     pool_config:
	//       max_open_conns: 100
	//       max_idle_conns: 10
	//       conn_max_lifetime: 3600  # 秒
	//       conn_max_idle_time: 600   # 秒
	//
	// 或者使用 PostgreSQL 数据库：
	//
	// database:
	//   driver: postgresql
	//   postgresql_config:
	//     dsn: "host=localhost port=5432 user=postgres password=password dbname=mydb sslmode=disable"
	//     pool_config:
	//       max_open_conns: 50
	//       max_idle_conns: 10
	//
	// 或者使用 SQLite 数据库：
	//
	// database:
	//   driver: sqlite
	//   sqlite_config:
	//     dsn: "file:./cache.db?cache=shared&mode=rwc"
}

// Example_databasemgr_basicOperations 演示基本的数据库操作
func Example_databasemgr_basicOperations() {
	// 假设已经创建了 dbMgr
	var dbMgr databasemgr.IDatabaseManager

	// 1. 简单查询
	// type User struct {
	//     ID   uint
	//     Name string
	// }
	// var user User
	// dbMgr.DB().First(&user, 1)

	// 2. 创建记录
	// user := User{Name: "John"}
	// dbMgr.DB().Create(&user)

	// 3. 更新记录
	// dbMgr.DB().Model(&user).Update("Name", "Jane")

	// 4. 删除记录
	// dbMgr.DB().Delete(&user)

	// 5. 事务操作
	// err := dbMgr.Transaction(func(tx *gorm.DB) error {
	//     if err := tx.Create(&User{Name: "Alice"}).Error; err != nil {
	//         return err  // 会回滚
	//     }
	//     return nil  // 提交
	// })

	_ = dbMgr
}

// Example_databasemgr_advancedOperations 演示高级数据库操作
func Example_databasemgr_advancedOperations() {
	// 假设已经创建了 dbMgr
	var dbMgr databasemgr.IDatabaseManager

	// 1. 自动迁移
	// err := dbMgr.AutoMigrate(&User{}, &Product{}, &Order{})

	// 2. 原生 SQL 查询
	// var results []map[string]interface{}
	// dbMgr.Raw("SELECT * FROM users WHERE age > ?", 18).Scan(&results)

	// 3. 执行原生 SQL
	// dbMgr.Exec("UPDATE users SET status = ? WHERE id = ?", "active", 1)

	// 4. 使用上下文（支持超时控制）
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// dbMgr.WithContext(ctx).Find(&users)

	// 5. 获取连接池统计信息
	// stats := dbMgr.Stats()
	// fmt.Printf("OpenConnections: %d\n", stats.OpenConnections)
	// fmt.Printf("InUse: %d\n", stats.InUse)

	// 6. 健康检查
	// if err := dbMgr.Health(); err != nil {
	//     loggerMgr := loggermgr.GetLoggerManager()
	//     loggerMgr.Logger("main").Error("Database health check failed", "error", err)
	// }

	_ = dbMgr
}

package databasemgr_test

import (
	"com.litelake.litecore/manager/databasemgr"
)

// Example_databasemgr_withDependencyInjection 演示如何使用依赖注入模式创建 DatabaseManager
func Example_databasemgr_withDependencyInjection() {
	// 创建数据库管理器（新方式 - 依赖注入）
	_ = databasemgr.NewManager("primary")

	// 将管理器注册到容器
	// container.Register(dbMgr)

	// 执行依赖注入
	// container.InjectAll()

	// 启动管理器（会自动从 ConfigProvider 加载配置）
	// if err := dbMgr.OnStart(); err != nil {
	//     log.Fatal(err)
	// }

	// 使用数据库
	// db := dbMgr.DB()
	// var users []User
	// db.Find(&users)

	// 优雅关闭
	// dbMgr.OnStop()
}

// Example_databasemgr_configuration 演示配置格式
func Example_databasemgr_configuration() {
	// 配置文件示例（YAML 格式）：
	//
	// database:
	//   primary:
	//     driver: mysql
	//     mysql_config:
	//       dsn: "root:password@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local"
	//       pool_config:
	//         max_open_conns: 100
	//         max_idle_conns: 10
	//         conn_max_lifetime: 3600  # 秒
	//         conn_max_idle_time: 600   # 秒
	//
	//   replica:
	//     driver: postgresql
	//     postgresql_config:
	//       dsn: "host=localhost port=5432 user=postgres password=password dbname=mydb sslmode=disable"
	//       pool_config:
	//         max_open_conns: 50
	//         max_idle_conns: 10
	//
	//   cache:
	//     driver: sqlite
	//     sqlite_config:
	//       dsn: "file:./cache.db?cache=shared&mode=rwc"
}

// Example_databasemgr_basicOperations 演示基本的数据库操作
func Example_databasemgr_basicOperations() {
	// 假设已经通过依赖注入初始化了 dbMgr
	var dbMgr *databasemgr.Manager

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
	// 假设已经通过依赖注入初始化了 dbMgr
	var dbMgr *databasemgr.Manager

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
	//     log.Error("database health check failed", err)
	// }

	_ = dbMgr
}

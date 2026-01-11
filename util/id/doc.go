// Package id 生成 CUID2 风格的分布式唯一标识符
//
// 核心特性：
//   - 时间有序：ID 前缀包含毫秒级时间戳，大致按时间排序
//   - 高唯一性：结合时间戳和加密级随机数，碰撞概率极低
//   - URL 安全：仅包含小写字母和数字，无特殊字符
//   - 分布式友好：无需中央协调，各节点独立生成
//   - 固定长度：标准 25 字符，便于数据库存储
//   - 高性能：单次生成耗时微秒级，支持高并发
//
// 基本用法：
//
//	id := id.NewCUID2()
//	fmt.Println(id) // 输出: 2k4d2j3h8f9g3n7p6q5r4s3t (示例)
//
// 在分布式系统中生成唯一主键：
//
//	type User struct {
//	    ID   string `gorm:"primaryKey"`
//	    Name string
//	}
//
//	user := &User{
//	    ID:   id.NewCUID2(),
//	    Name: "Alice",
//	}
//
// 批量生成唯一标识符：
//
//	ids := make([]string, 1000)
//	for i := range ids {
//	    ids[i] = id.NewCUID2()
//	}
//
// CUID2 特性说明：
//   - 编码方式：Base36 编码（0-9 和 a-z），比十六进制更紧凑
//   - 时间戳：使用 Unix 毫秒时间戳作为前缀，保证时间排序性
//   - 随机性：使用 crypto/rand 生成 16 字节加密级随机数
//   - 唯一性保证：理论碰撞概率远低于 UUID v4
//   - 适用场景：数据库主键、分布式追踪 ID、会话 ID、订单号等
package id

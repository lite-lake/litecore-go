package server

// 引入 time/tzdata，将 IANA 时区数据库嵌入二进制。
// Windows 不自带 zoneinfo 数据库，不导入会导致 time.LoadLocation("Asia/Shanghai") 等调用失败。
import _ "time/tzdata"

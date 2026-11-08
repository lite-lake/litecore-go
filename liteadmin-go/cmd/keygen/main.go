// keygen 命令行工具：生成一把 Admin 委托密钥（32 字节随机数，base64url 编码）。
//
// 用法：
//
//	go run ./cmd/keygen          # 输出一把新 key
//	go run ./cmd/keygen -n 3     # 一次输出 3 把（轮换时新旧并存）
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lite-lake/liteadmin-go"
)

func main() {
	count := flag.Int("n", 1, "生成数量")
	flag.Parse()
	if *count < 1 {
		fmt.Fprintln(os.Stderr, "生成数量必须 ≥ 1")
		os.Exit(1)
	}
	for i := 0; i < *count; i++ {
		key, err := liteadmin.GenerateKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "生成密钥失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(key)
	}
}

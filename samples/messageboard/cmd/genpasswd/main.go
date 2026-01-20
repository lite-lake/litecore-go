// Package main 提供管理员密码生成工具
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"com.litelake.litecore/util/hash"
)

func main() {
	fmt.Println("=== 留言板管理员密码生成工具 ===")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("请输入管理员密码: ")
		password, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取失败: %v\n", err)
			os.Exit(1)
		}

		password = strings.TrimSpace(password)
		if password == "" {
			fmt.Println("密码不能为空，请重新输入")
			continue
		}

		fmt.Print("再次输入密码确认: ")
		confirmPassword, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取失败: %v\n", err)
			os.Exit(1)
		}

		confirmPassword = strings.TrimSpace(confirmPassword)

		if password != confirmPassword {
			fmt.Println("两次输入的密码不一致，请重新输入")
			continue
		}

		hashedPassword, err := hash.Hash.BcryptHash(password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加密失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println("=== 生成成功 ===")
		fmt.Printf("加密后的密码: %s\n", hashedPassword)
		fmt.Println()
		fmt.Println("请将上面的加密密码复制到 config.yaml 文件的 app.admin.password 字段")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  app:")
		fmt.Println("    admin:")
		fmt.Println("      password: \"$2a$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\"")
		fmt.Println()

		fmt.Print("是否继续生成其他密码? (y/N): ")
		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取失败: %v\n", err)
			os.Exit(1)
		}

		choice = strings.ToLower(strings.TrimSpace(choice))
		if choice != "y" && choice != "yes" {
			break
		}
		fmt.Println()
	}

	fmt.Println("退出程序")
}

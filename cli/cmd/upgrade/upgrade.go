package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/lite-lake/litecore-go/cli/internal/version"
	"github.com/urfave/cli/v3"
)

const (
	githubRepo   = "lite-lake/litecore-go"
	githubAPIURL = "https://api.github.com/repos/%s/releases/latest"
	goInstallURL = "github.com/%s/cli@%s"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func GetCommand() *cli.Command {
	return &cli.Command{
		Name:        "upgrade",
		Usage:       "升级 CLI 到最新版本",
		Description: `从 GitHub 检查并安装最新版本的 LiteCore CLI`,
		Action:      runUpgrade,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "force",
				Usage: "强制升级，即使已经是最新版本",
			},
			&cli.BoolFlag{
				Name:  "check",
				Usage: "仅检查是否有新版本，不执行升级",
			},
		},
	}
}

func runUpgrade(ctx context.Context, cmd *cli.Command) error {
	fmt.Printf("当前版本: %s\n", version.Version)

	latestRelease, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("获取最新版本失败: %w", err)
	}

	latestVersion := latestRelease.TagName
	fmt.Printf("最新版本: %s\n", latestVersion)

	if cmd.Bool("check") {
		if latestVersion == version.Version {
			fmt.Println("已经是最新版本")
		} else {
			fmt.Printf("有新版本可用: %s\n", latestRelease.HTMLURL)
		}
		return nil
	}

	if latestVersion == version.Version && !cmd.Bool("force") {
		fmt.Println("已经是最新版本，使用 --force 强制升级")
		return nil
	}

	fmt.Println("正在升级...")

	if err := installLatestVersion(latestVersion); err != nil {
		return fmt.Errorf("升级失败: %w", err)
	}

	fmt.Println("升级成功!")
	return nil
}

func getLatestRelease() (*githubRelease, error) {
	url := fmt.Sprintf(githubAPIURL, githubRepo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

func installLatestVersion(targetVersion string) error {
	installURL := fmt.Sprintf(goInstallURL, githubRepo, targetVersion)

	goBin, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("未找到 go 命令: %w", err)
	}

	installCmd := exec.Command(goBin, "install", installURL)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installCmd.Env = os.Environ()

	if err := installCmd.Run(); err != nil {
		if isWindowsFileLocked(err) {
			return handleWindowsFileLock(targetVersion, goBin, installURL)
		}
		return err
	}

	return nil
}

func isWindowsFileLocked(err error) bool {
	if runtime.GOOS != "windows" {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "access is denied") ||
		strings.Contains(errMsg, "used by another process") ||
		strings.Contains(errMsg, "being used")
}

func handleWindowsFileLock(targetVersion, goBin, installURL string) error {
	fmt.Println("\n检测到文件被占用，尝试使用临时目录方式升级...")

	tempDir, err := os.MkdirTemp("", "litecore-upgrade-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	oldGoBin := os.Getenv("GOBIN")
	os.Setenv("GOBIN", tempDir)
	defer func() {
		if oldGoBin != "" {
			os.Setenv("GOBIN", oldGoBin)
		} else {
			os.Unsetenv("GOBIN")
		}
	}()

	installCmd := exec.Command(goBin, "install", installURL)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installCmd.Env = os.Environ()

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("临时安装失败: %w", err)
	}

	execName := "litecore-cli"
	if runtime.GOOS == "windows" {
		execName = "litecore-cli.exe"
	}

	tempBinary := tempDir + string(os.PathSeparator) + execName
	targetBinary := getTargetBinaryPath()

	if _, err := os.Stat(tempBinary); os.IsNotExist(err) {
		return fmt.Errorf("临时二进制文件不存在: %s", tempBinary)
	}

	backupPath := targetBinary + ".old"
	if _, err := os.Stat(targetBinary); err == nil {
		os.Rename(targetBinary, backupPath)
		defer os.Remove(backupPath)
	}

	if err := copyFile(tempBinary, targetBinary); err != nil {
		if err := os.Rename(backupPath, targetBinary); err != nil {
			fmt.Printf("恢复备份失败: %v\n", err)
		}
		return fmt.Errorf("复制新版本失败: %w", err)
	}

	return nil
}

func getTargetBinaryPath() string {
	gobin := os.Getenv("GOBIN")
	if gobin != "" {
		return gobin + string(os.PathSeparator) + getBinaryName()
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		homeDir, _ := os.UserHomeDir()
		gopath = homeDir + string(os.PathSeparator) + "go"
	}

	return gopath + string(os.PathSeparator) + "bin" + string(os.PathSeparator) + getBinaryName()
}

func getBinaryName() string {
	if runtime.GOOS == "windows" {
		return "litecore-cli.exe"
	}
	return "litecore-cli"
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

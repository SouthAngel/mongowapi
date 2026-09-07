//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const serviceName = "mongowapi"
const serviceDesc = "MongoDB RESTful Web API"

// systemd unit 文件路径
const unitFile = "/etc/systemd/system/mongowapi.service"

// isServiceMode 检测是否由 systemd 启动（PID 1 为 init/systemd）
func isServiceMode() bool {
	ppid := os.Getppid()
	if ppid == 1 {
		return true
	}
	// 检查是否在 systemd 环境下运行
	if os.Getenv("INVOCATION_ID") != "" {
		return true
	}
	return false
}

// exePath 获取当前可执行文件的绝对路径
func exePath() (string, error) {
	p, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	return filepath.Abs(p)
}

// installService 生成 systemd unit 文件并注册服务
func installService(configPath string) error {
	exe, err := exePath()
	if err != nil {
		return err
	}
	absConfig, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}

	// 生成 systemd unit 文件
	unit := fmt.Sprintf(`[Unit]
Description=%s
After=network.target

[Service]
Type=simple
ExecStart=%s -config %s
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`, serviceDesc, exe, absConfig)

	// 写入 unit 文件（需要 root 权限）
	if err := os.WriteFile(unitFile, []byte(unit), 0644); err != nil {
		return fmt.Errorf("写入 unit 文件失败（需要 root 权限）: %w", err)
	}

	// 重新加载 systemd 配置
	if err := runCmd("systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("systemctl daemon-reload 失败: %w", err)
	}

	// 启用服务（开机自启）
	if err := runCmd("systemctl", "enable", serviceName); err != nil {
		return fmt.Errorf("systemctl enable 失败: %w", err)
	}

	return nil
}

// uninstallService 移除 systemd 服务
func uninstallService() error {
	// 先停止服务（忽略错误，可能未运行）
	_ = runCmd("systemctl", "stop", serviceName)

	// 禁用服务
	if err := runCmd("systemctl", "disable", serviceName); err != nil {
		return fmt.Errorf("systemctl disable 失败: %w", err)
	}

	// 删除 unit 文件
	if err := os.Remove(unitFile); err != nil {
		return fmt.Errorf("删除 unit 文件失败: %w", err)
	}

	// 重新加载 systemd 配置
	if err := runCmd("systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("systemctl daemon-reload 失败: %w", err)
	}

	return nil
}

// startService 启动 systemd 服务
func startService() error {
	return runCmd("systemctl", "start", serviceName)
}

// stopService 停止 systemd 服务
func stopService() error {
	return runCmd("systemctl", "stop", serviceName)
}

// runAsService 以 systemd 服务模式运行（systemd 通过 SIGTERM 发送停止信号）
func runAsService(configPath string) {
	runForeground(configPath)
}

// runCmd 执行系统命令
func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
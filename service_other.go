//go:build !windows && !linux

package main

import "fmt"

const serviceName = "mongowapi"

func isServiceMode() bool {
	return false
}

func installService(configPath string) error {
	return fmt.Errorf("当前平台不支持服务注册")
}

func uninstallService() error {
	return fmt.Errorf("当前平台不支持服务移除")
}

func startService() error {
	return fmt.Errorf("当前平台不支持服务管理")
}

func stopService() error {
	return fmt.Errorf("当前平台不支持服务管理")
}

func runAsService(configPath string) {
	runForeground(configPath)
}

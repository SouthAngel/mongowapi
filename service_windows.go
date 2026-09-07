//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "MongoWAPI"
const serviceDesc = "MongoDB RESTful Web API"

// isServiceMode 检测是否由 Windows SCM 启动
func isServiceMode() bool {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return false
	}
	return isService
}

// installService 注册为 Windows 服务
func installService(configPath string) error {
	exe, err := exePath()
	if err != nil {
		return err
	}
	absConfig, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接 SCM 失败: %w", err)
	}
	defer m.Disconnect()

	// 若服务已存在则先删除
	if s, err := m.OpenService(serviceName); err == nil {
		s.Close()
		_ = doUninstall()
	}

	s, err := m.CreateService(serviceName, exe, mgr.Config{
		DisplayName: serviceName,
		Description: serviceDesc,
		StartType:   mgr.StartAutomatic,
	}, "-config", absConfig)
	if err != nil {
		return fmt.Errorf("创建服务失败: %w", err)
	}
	defer s.Close()

	return nil
}

// doUninstall 实际移除逻辑（不导出，供 installService 和 uninstallService 复用）
func doUninstall() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接 SCM 失败: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("服务不存在: %w", err)
	}
	defer s.Close()

	// 先停止再删除
	_, _ = s.Control(svc.Stop)
	return s.Delete()
}

// uninstallService 移除 Windows 服务
func uninstallService() error {
	return doUninstall()
}

// startService 启动 Windows 服务
func startService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接 SCM 失败: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("服务不存在: %w", err)
	}
	defer s.Close()

	return s.Start()
}

// stopService 停止 Windows 服务
func stopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("连接 SCM 失败: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("服务不存在: %w", err)
	}
	defer s.Close()

	_, err = s.Control(svc.Stop)
	return err
}

// runAsService 以 Windows 服务模式运行
func runAsService(configPath string) {
	s := &winService{configPath: configPath}
	if err := svc.Run(serviceName, s); err != nil {
		log.Fatalf("服务运行失败: %v", err)
	}
}

// winService 实现 svc.Handler 接口
type winService struct {
	configPath string
}

// Execute 是 svc.Handler 的实现，服务启动时由 SCM 调用
func (s *winService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	// 上报服务正在运行
	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	// 异步启动应用，通过 channel 返回 app 供停止时优雅关闭
	appCh := make(chan *app, 1)
	go func() {
		a, err := startApp(s.configPath)
		if err != nil {
			log.Printf("启动应用失败: %v", err)
			appCh <- nil
			return
		}
		appCh <- a
	}()

	// 监听服务控制请求
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}

				// 等待应用启动完成并优雅关闭
				select {
				case a := <-appCh:
					if a != nil {
						a.shutdown()
					}
				case <-time.After(5 * time.Second):
					log.Println("等待应用启动超时，跳过优雅关闭")
				}

				changes <- svc.Status{State: svc.Stopped}
				return
			default:
				// 忽略其他控制请求
			}
		}
	}
}

// exePath 获取当前可执行文件的绝对路径
func exePath() (string, error) {
	p, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	return filepath.Abs(p)
}
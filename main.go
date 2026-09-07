package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	install := flag.Bool("install", false, "注册为系统服务")
	uninstall := flag.Bool("uninstall", false, "移除系统服务")
	start := flag.Bool("start", false, "启动系统服务")
	stop := flag.Bool("stop", false, "停止系统服务")
	flag.Parse()

	// 服务管理命令
	switch {
	case *install:
		if err := installService(*configPath); err != nil {
			log.Fatalf("注册服务失败: %v", err)
		}
		log.Println("服务注册成功")
		return
	case *uninstall:
		if err := uninstallService(); err != nil {
			log.Fatalf("移除服务失败: %v", err)
		}
		log.Println("服务移除成功")
		return
	case *start:
		if err := startService(); err != nil {
			log.Fatalf("启动服务失败: %v", err)
		}
		log.Println("服务已启动")
		return
	case *stop:
		if err := stopService(); err != nil {
			log.Fatalf("停止服务失败: %v", err)
		}
		log.Println("服务已停止")
		return
	}

	// 服务模式运行检测
	if isServiceMode() {
		runAsService(*configPath)
		return
	}

	// 普通前台运行
	runForeground(*configPath)
}

// runForeground 前台运行，等待中断信号后优雅关闭
func runForeground(configPath string) {
	a, err := startApp(configPath)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// 等待中断/终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("收到停止信号，正在关闭服务...")
	a.shutdown()
}
// Package logger 提供基于 lumberjack 的日志输出管理。
// 支持同时输出到控制台与文件，文件按大小自动滚动并按天保留旧日志。
package logger

import (
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"

	"mongowapi/config"
)

// Setup 根据日志配置初始化标准 log 与 Gin 的日志输出。
// 当 config.Log.Filename 非空时写入文件（lumberjack 滚动），否则仅输出到 stdout。
// 返回的关闭函数用于刷新并关闭日志文件，应在应用退出时调用。
func Setup(cfg *config.LogConfig) (close func() error) {
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	if cfg.Filename != "" {
		lj := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
			LocalTime:  cfg.LocalTime,
		}
		writers = append(writers, lj)
		log.SetOutput(io.MultiWriter(writers...))

		gin.DefaultWriter = io.MultiWriter(writers...)
		gin.DefaultErrorWriter = io.MultiWriter(writers...)

		return lj.Close
	}

	log.SetOutput(io.MultiWriter(writers...))
	gin.DefaultWriter = io.MultiWriter(writers...)
	gin.DefaultErrorWriter = io.MultiWriter(writers...)

	return func() error { return nil }
}
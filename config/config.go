package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	Server ServerConfig `yaml:"server"`
	Mongo  MongoConfig  `yaml:"mongo"`
	Log    LogConfig    `yaml:"log"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

// MongoConfig MongoDB 连接配置
type MongoConfig struct {
	URI            string        `yaml:"uri"`
	ConnectTimeout time.Duration `yaml:"connect_timeout"`
	RequestTimeout  time.Duration `yaml:"request_timeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	Filename   string `yaml:"filename"`    // 日志文件路径（为空则不写文件）
	MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小（MB）
	MaxAge     int    `yaml:"max_age"`     // 保留的旧日志最大天数
	MaxBackups int    `yaml:"max_backups"` // 保留的旧日志最大数量
	Compress   bool   `yaml:"compress"`    // 是否压缩旧日志
	LocalTime  bool   `yaml:"local_time"`  // 是否使用本地时间命名备份文件
}

// Load 从指定路径加载配置文件
// 支持 YAML 和 JSON 两种格式（YAML 为 JSON 的超集）
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	cfg.setDefaults()
	return cfg, nil
}

func (c *Config) setDefaults() {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "debug"
	}
	if c.Mongo.URI == "" {
		c.Mongo.URI = "mongodb://localhost:27017"
	}
	if c.Mongo.ConnectTimeout == 0 {
		c.Mongo.ConnectTimeout = 10 * time.Second
	}
	if c.Mongo.RequestTimeout == 0 {
		c.Mongo.RequestTimeout = 30 * time.Second
	}
	if c.Log.MaxSize == 0 {
		c.Log.MaxSize = 100 // 默认 100MB
	}
	if c.Log.MaxAge == 0 {
		c.Log.MaxAge = 7 // 默认保留 7 天
	}
}
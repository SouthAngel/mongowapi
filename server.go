package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"mongowapi/config"
	"mongowapi/database"
	"mongowapi/handlers"
	"mongowapi/logger"
)

// app 封装 HTTP 服务与 MongoDB 资源，供前台与服务模式复用
type app struct {
	srv       *http.Server
	mdb       *database.Mongo
	cfg       *config.Config
	closeLog  func() error
}

// startApp 加载配置、连接 MongoDB 并启动 HTTP 服务（非阻塞）
func startApp(configPath string) (*app, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 初始化日志输出（控制台 + 可选文件）
	closeLog := logger.Setup(&cfg.Log)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	gin.SetMode(cfg.Server.Mode)

	ctx := context.Background()
	mdb, err := database.New(ctx, cfg.Mongo.URI, cfg.Mongo.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	log.Println("MongoDB 连接成功")

	r := gin.Default()
	registerRoutes(r, mdb, cfg)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 服务异常退出: %v", err)
		}
	}()
	log.Printf("HTTP 服务已启动，监听 %s", srv.Addr)

	return &app{srv: srv, mdb: mdb, cfg: cfg, closeLog: closeLog}, nil
}

// registerRoutes 注册所有 REST 路由（统一 POST + 动作端点，参数全部在请求体）
func registerRoutes(r *gin.Engine, mdb *database.Mongo, cfg *config.Config) {
	h := handlers.NewHandler(mdb, cfg.Mongo.RequestTimeout, cfg.Mongo.WhiteListDB)

	api := r.Group("/api")
	{
		// 文档 CRUD
		api.POST("/insert", h.InsertOne)      // 插入单条
		api.POST("/insertmany", h.InsertMany) // 批量插入
		api.POST("/find", h.Find)             // 查询列表
		api.POST("/findone", h.FindOne)       // 查询单条
		api.POST("/update", h.UpdateOne)      // 更新单条
		api.POST("/updatemany", h.UpdateMany) // 批量更新
		api.POST("/delete", h.DeleteOne)      // 删除单条
		api.POST("/deletemany", h.DeleteMany) // 批量删除

		// 索引管理
		api.POST("/indexes", h.ListIndexes)    // 列出索引
		api.POST("/createindex", h.CreateIndex) // 创建索引
		api.POST("/dropindex", h.DropIndex)     // 删除索引
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

// shutdown 优雅关闭 HTTP 服务并释放 MongoDB 连接
func (a *app) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Mongo.RequestTimeout)
	defer cancel()

	if err := a.srv.Shutdown(ctx); err != nil {
		log.Printf("关闭 HTTP 服务失败: %v", err)
	}
	if err := a.mdb.Close(ctx); err != nil {
		log.Printf("关闭 MongoDB 连接失败: %v", err)
	}
	if err := a.closeLog(); err != nil {
		log.Printf("关闭日志文件失败: %v", err)
	}
}
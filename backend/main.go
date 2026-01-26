package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"io.lazydoge/aclove/config"
	"io.lazydoge/aclove/database"
	"io.lazydoge/aclove/handlers"
	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/repository"
	"io.lazydoge/aclove/routes"
	"io.lazydoge/aclove/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	logger.Init("INFO")

	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Fatal("加载配置失败", "error", err)
	}

	gormDB, err := database.NewGORM(&cfg.Database)
	if err != nil {
		logger.Fatal("连接数据库失败", "error", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		logger.Fatal("获取数据库实例失败", "error", err)
	}
	defer sqlDB.Close()

	if err := initDatabase(gormDB); err != nil {
		logger.Fatal("初始化数据库失败", "error", err)
	}

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	categoryRepo := repository.NewCategoryRepository(gormDB)
	categorySvc := service.NewCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categorySvc)

	routes.Setup(router, categoryHandler)

	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	logger.Info("服务器启动", "address", addr)

	go func() {
		if err := router.Run(addr); err != nil {
			logger.Fatal("启动服务器失败", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func initDatabase(db *gorm.DB) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC);
	`
	return db.Exec(createTableSQL).Error
}

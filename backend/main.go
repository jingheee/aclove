package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"io.lazydoge/aclove/config"
	"io.lazydoge/aclove/database"
	io "io.lazydoge/aclove/handlers"
	"io.lazydoge/aclove/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	if err := initDatabase(db); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	router := gin.Default()

	messageHandler := io.NewMessageHandler(db)
	routes.Setup(router, messageHandler)

	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	log.Printf("服务器启动在 %s", addr)

	go func() {
		if err := router.Run(addr); err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")
}

func initDatabase(db *database.DB) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC);
	`
	_, err := db.Exec(createTableSQL)
	return err
}

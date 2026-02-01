package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/rueidis"
	"gorm.io/gorm"

	"io.lazydoge/aclove/cache"
	"io.lazydoge/aclove/config"
	"io.lazydoge/aclove/database"
	"io.lazydoge/aclove/handlers"
	"io.lazydoge/aclove/jsonutil"
	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/models/query"
	"io.lazydoge/aclove/repository"
	"io.lazydoge/aclove/routes"
	"io.lazydoge/aclove/service"
	"io.lazydoge/aclove/session"
	"io.lazydoge/aclove/storage/minio"
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

	jsonutil.EnableCustomJSONBinding()
	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Device-Fingerprint"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig))

	categoryRepo := repository.NewCategoryRepository(gormDB)
	categorySvc := service.NewCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categorySvc)

	anonymousUserRepo := query.NewAnonymousUserRepo(gormDB)
	anonymousUserSvc := service.NewAnonymousUserService(anonymousUserRepo)
	anonymousUserHandler := handlers.NewAnonymousUserHandler(anonymousUserSvc)

	postRepo := query.NewPostRepo(gormDB)

	var categoryCache *cache.Cache
	var sessionManager *session.Manager
	var redisClient rueidis.Client

	if cfg.Redis.Addr != "" {
		categoryCache, err = cache.New(cache.Config{
			Addr:        cfg.Redis.Addr,
			Password:    cfg.Redis.Password,
			DB:          cfg.Redis.DB,
			Prefix:      cfg.Redis.CachePrefix,
			DefaultTTLD: cfg.Redis.CacheTTL,
		})
		if err != nil {
			logger.Warn("连接Redis失败，缓存功能不可用", "error", err)
			categoryCache = nil
		}

		redisClient, err = rueidis.NewClient(rueidis.ClientOption{
			InitAddress: []string{cfg.Redis.Addr},
			Password:    cfg.Redis.Password,
			SelectDB:    cfg.Redis.DB,
		})
		if err != nil {
			logger.Warn("连接Redis客户端失败", "error", err)
		}

		sessionManager, err = session.NewManager(anonymousUserRepo, session.ManagerConfig{
			StoreConfig: session.StoreConfig{
				RedisAddr:       cfg.Redis.Addr,
				RedisPassword:   cfg.Redis.Password,
				RedisDB:         cfg.Redis.DB,
				KeyPrefix:       cfg.Redis.SessionPrefix,
				SessionTTL:      cfg.Redis.SessionTTL,
				CleanupInterval: cfg.Redis.SessionCleanupInterval,
			},
			CookieName:     cfg.Session.CookieName,
			CookieDomain:   cfg.Session.CookieDomain,
			CookieSecure:   cfg.Session.CookieSecure,
			CookieHttpOnly: cfg.Session.CookieHttpOnly,
			CookieSameSite: cfg.Session.CookieSameSite,
		})
		if err != nil {
			logger.Fatal("创建Session Manager失败", "error", err)
		}
		defer sessionManager.Close()
	} else {
		logger.Fatal("Redis配置不能为空，分布式session需要Redis支持")
	}

	postSvc := service.NewPostService(postRepo, categoryRepo, redisClient)
	postHandler := handlers.NewPostHandler(postSvc)

	// Initialize MinIO storage
	minioClient, err := minio.NewClient(minio.Config{
		Endpoint:        cfg.Storage.MinIO.Endpoint,
		Bucket:          cfg.Storage.MinIO.Bucket,
		AccessKeyID:     cfg.Storage.MinIO.AccessKeyID,
		SecretAccessKey: cfg.Storage.MinIO.SecretAccessKey,
		Region:          cfg.Storage.MinIO.Region,
		PublicURL:       cfg.Storage.MinIO.PublicURL,
	})
	if err != nil {
		logger.Fatal("创建MinIO客户端失败", "error", err)
	}
	storageService := minio.NewService(minio.ServiceConfig{
		Client:        minioClient,
		MaxFileSize:   cfg.Storage.MinIO.MaxFileSize,
		MaxConcurrent: cfg.Storage.MinIO.MaxConcurrent,
	})
	uploadHandler := handlers.NewUploadHandler(storageService)

	routes.Setup(router, categoryHandler, categoryCache, anonymousUserSvc, anonymousUserHandler, sessionManager, postHandler, uploadHandler)

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

	logger.Info("服务器正在关闭")
}

func initDatabase(db *gorm.DB) error {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC);

		CREATE TABLE IF NOT EXISTS anonymous_users (
			id BIGSERIAL PRIMARY KEY,
			cookie VARCHAR(64) NOT NULL UNIQUE,
			fingerprint_hash VARCHAR(128),
			ip VARCHAR(45) NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			status_reason TEXT,
			banned_until TIMESTAMP WITH TIME ZONE,
			cooldown_until TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP WITH TIME ZONE
		);

		CREATE INDEX IF NOT EXISTS idx_anonymous_users_cookie ON anonymous_users(cookie);
		CREATE INDEX IF NOT EXISTS idx_anonymous_users_status ON anonymous_users(status);
		CREATE INDEX IF NOT EXISTS idx_anonymous_users_deleted_at ON anonymous_users(deleted_at);
	`
	return db.Exec(createTableSQL).Error
}

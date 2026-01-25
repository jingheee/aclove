package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"io.lazydoge/aclove/config"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func New(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetimeDuration())

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("测试数据库连接失败: %w", err)
	}

	return &DB{db}, nil
}

func (db *DB) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

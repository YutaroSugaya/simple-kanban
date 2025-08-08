package repository

import (
	"fmt"
	"log"

	"simple-kanban/config"
	"simple-kanban/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB データベース接続のグローバル変数
var DB *gorm.DB

// InitDB データベース接続を初期化します
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// データベース接続文字列を生成
	dsn := cfg.GetDSN()

	// GORMのログレベルを設定
	var logLevel logger.LogLevel
	switch cfg.Server.Mode {
	case "release":
		logLevel = logger.Error
	case "test":
		logLevel = logger.Silent
	default:
		logLevel = logger.Info
	}

	// データベースに接続
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("データベース接続に失敗しました: %w", err)
	}

	// 接続プールの設定
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("データベースインスタンス取得に失敗しました: %w", err)
	}

	// 接続プールの設定
	sqlDB.SetMaxIdleConns(10)  // アイドル接続の最大数
	sqlDB.SetMaxOpenConns(100) // オープン接続の最大数

	// グローバル変数に設定
	DB = db

	log.Println("データベース接続が正常に確立されました")
	return db, nil
}

// Migrate データベースのマイグレーションを実行します
func Migrate(db *gorm.DB) error {
	log.Println("データベースマイグレーションを開始します...")

	// すべてのエンティティのマイグレーションを実行
	err := db.AutoMigrate(
		&domain.User{},
		&domain.Board{},
		&domain.Column{},
		&domain.Task{},
		&domain.CalendarSettings{},
		&domain.TimerSession{},
		&domain.CalendarEvent{},
	)
	if err != nil {
		return fmt.Errorf("マイグレーションに失敗しました: %w", err)
	}

	log.Println("データベースマイグレーションが完了しました")
	return nil
}

// CloseDB データベース接続を閉じます
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB データベース接続を取得します
func GetDB() *gorm.DB {
	return DB
}

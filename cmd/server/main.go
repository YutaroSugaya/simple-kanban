package main

import (
	"log"
	"net/http"

	"simple-kanban/config"
	"simple-kanban/internal/handler"
	"simple-kanban/internal/repository"
	"simple-kanban/internal/service"
	"simple-kanban/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 設定を読み込み
	cfg := config.Load()
	log.Printf("アプリケーション設定を読み込みました: DB=%s:%d, Port=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Server.Port)

	// Ginのモードを設定
	gin.SetMode(cfg.Server.Mode)

	// データベースに接続
	db, err := repository.InitDB(cfg)
	if err != nil {
		log.Fatalf("データベース接続エラー: %v", err)
	}
	defer func() {
		if err := repository.CloseDB(); err != nil {
			log.Printf("データベース接続終了エラー: %v", err)
		}
	}()

	// データベースマイグレーションを実行
	if err := repository.Migrate(db); err != nil {
		log.Fatalf("データベースマイグレーションエラー: %v", err)
	}

	// リポジトリレイヤーを初期化
	userRepo := repository.NewUserRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	columnRepo := repository.NewColumnRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// サービスレイヤーを初期化
	userService := service.NewUserService(userRepo, cfg)
	boardService := service.NewBoardService(boardRepo, db)
	taskService := service.NewTaskService(taskRepo, boardRepo, columnRepo)

	// ハンドラーレイヤーを初期化
	authHandler := handler.NewAuthHandler(userService)
	boardHandler := handler.NewBoardHandler(boardService)
	taskHandler := handler.NewTaskHandler(taskService)

	// Ginルーターを作成
	router := gin.New()

	// ミドルウェアを設定
	router.Use(gin.Logger())   // リクエストログを出力
	router.Use(gin.Recovery()) // パニック時の復旧

	// CORS設定（開発用）
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// ヘルスチェックエンドポイント
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "simple-kanban-api",
		})
	})

	// API v1 ルートグループ
	v1 := router.Group("/api/v1")
	{
		// 認証エンドポイント（認証不要）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register) // ユーザー登録
			auth.POST("/login", authHandler.Login)       // ログイン
		}

		// 認証が必要なエンドポイント
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg)) // JWT認証ミドルウェア
		{
			// 認証関連（認証後）
			protected.GET("/auth/profile", authHandler.Profile) // プロフィール取得

			// ボード関連
			boards := protected.Group("/boards")
			{
				boards.GET("", boardHandler.GetUserBoards)                   // ボード一覧取得
				boards.POST("", boardHandler.CreateBoard)                    // ボード作成
				boards.GET("/:id/columns", boardHandler.GetBoardWithColumns) // ボード詳細（カラム付き）
				boards.PUT("/:id", boardHandler.UpdateBoard)                 // ボード更新
				boards.DELETE("/:id", boardHandler.DeleteBoard)              // ボード削除
			}

			// タスク関連
			tasks := protected.Group("/tasks")
			{
				tasks.POST("", taskHandler.CreateTask)       // タスク作成
				tasks.GET("/:id", taskHandler.GetTask)       // タスク取得
				tasks.PUT("/:id", taskHandler.UpdateTask)    // タスク更新
				tasks.DELETE("/:id", taskHandler.DeleteTask) // タスク削除
				tasks.PUT("/:id/move", taskHandler.MoveTask) // タスク移動
			}

			// カラム関連（タスクの順序変更）
			columns := protected.Group("/columns")
			{
				columns.PUT("/:columnId/tasks/reorder", taskHandler.ReorderTasks) // タスク順序変更
			}
		}
	}

	// サーバー起動
	log.Printf("サーバーをポート %s で起動します...", cfg.Server.Port)
	log.Printf("ヘルスチェック: http://localhost:%s/health", cfg.Server.Port)
	log.Printf("API エンドポイント: http://localhost:%s/api/v1", cfg.Server.Port)

	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("サーバー起動エラー: %v", err)
	}
}

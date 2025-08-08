package main

import (
	"log"
	"net/http"

	"simple-kanban/config"
	"simple-kanban/internal/handler"
	"simple-kanban/internal/repository"
	"simple-kanban/internal/service"
	"simple-kanban/pkg/logger"
	"simple-kanban/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// ロガーを初期化
	appLogger, err := logger.NewLogger("logs/app.log")
	if err != nil {
		log.Fatalf("ロガーの初期化に失敗しました: %v", err)
	}
	defer appLogger.Close()

	// 設定を読み込み
	cfg := config.Load()
	appLogger.Info("アプリケーション設定を読み込みました: DB=%s:%d, Port=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Server.Port)

	// Ginのモードを設定
	gin.SetMode(cfg.Server.Mode)

	// データベースに接続
	db, err := repository.InitDB(cfg)
	if err != nil {
		appLogger.Error("データベース接続エラー: %v", err)
		log.Fatalf("データベース接続エラー: %v", err)
	}
	defer func() {
		if err := repository.CloseDB(); err != nil {
			appLogger.Error("データベース接続終了エラー: %v", err)
		}
	}()

	// データベースマイグレーションを実行
	if err := repository.Migrate(db); err != nil {
		appLogger.Error("データベースマイグレーションエラー: %v", err)
		log.Fatalf("データベースマイグレーションエラー: %v", err)
	}

	// リポジトリレイヤーを初期化
	userRepo := repository.NewUserRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	columnRepo := repository.NewColumnRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	calendarSettingsRepo := repository.NewCalendarSettingsRepository(db)
	calendarEventRepo := repository.NewCalendarEventRepository(db)
	timerSessionRepo := repository.NewTimerSessionRepository(db)

	// サービスレイヤーを初期化
	userService := service.NewUserService(userRepo, cfg)
	boardService := service.NewBoardService(boardRepo, db)
	taskService := service.NewTaskService(taskRepo, boardRepo, columnRepo)
	calendarService := service.NewCalendarService(calendarSettingsRepo, calendarEventRepo, taskRepo)
	timerService := service.NewTimerService(timerSessionRepo, taskRepo)

	// ハンドラーレイヤーを初期化
	authHandler := handler.NewAuthHandler(userService)
	boardHandler := handler.NewBoardHandler(boardService)
	taskHandler := handler.NewTaskHandler(taskService)
	calendarHandler := handler.NewCalendarHandler(calendarService, taskService, appLogger)
	timerHandler := handler.NewTimerHandler(timerService)
	analyticsHandler := handler.NewAnalyticsHandler()

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
				boards.GET("", boardHandler.GetUserBoards)                         // ボード一覧取得
				boards.GET("/with-columns", boardHandler.GetUserBoardsWithColumns) // ボード一覧取得（カラム・タスク付き）
				boards.POST("", boardHandler.CreateBoard)                          // ボード作成
				boards.GET("/:id/columns", boardHandler.GetBoardWithColumns)       // ボード詳細（カラム付き）
				boards.PUT("/:id", boardHandler.UpdateBoard)                       // ボード更新
				boards.DELETE("/:id", boardHandler.DeleteBoard)                    // ボード削除
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

			// カレンダー関連
			calendar := protected.Group("/calendar")
			{
				calendar.GET("/settings", calendarHandler.GetCalendarSettings)          // カレンダー設定取得
				calendar.PUT("/settings", calendarHandler.UpdateCalendarSettings)       // カレンダー設定更新
				calendar.GET("/events", calendarHandler.GetEvents)                      // イベント取得
				calendar.POST("/events", calendarHandler.CreateEvent)                   // イベント作成
				calendar.PUT("/events/:id", calendarHandler.UpdateEvent)                // イベント更新
				calendar.DELETE("/events/:id", calendarHandler.DeleteEvent)             // イベント削除
				calendar.POST("/tasks/:taskId/events", calendarHandler.CreateTaskEvent) // タスクからイベント作成
			}

			// タイマー関連
			timer := protected.Group("/timer")
			{
				timer.POST("/start", timerHandler.StartTimer)             // タイマー開始
				timer.PUT("/:id/stop", timerHandler.StopTimer)            // タイマー停止
				timer.GET("/active", timerHandler.GetActiveTimer)         // アクティブタイマー取得
				timer.GET("/history", timerHandler.GetTimerHistory)       // タイマー履歴取得
				timer.GET("/tasks/:taskId", timerHandler.GetTimersByTask) // タスク別タイマー履歴
			}

			// 分析・統計関連
			analytics := protected.Group("/analytics")
			{
				analytics.GET("/task-completion", analyticsHandler.GetTaskCompletionStats) // タスク完了統計
			}
		}
	}

	// サーバー起動
	appLogger.Info("サーバーをポート %s で起動します...", cfg.Server.Port)
	appLogger.Info("ヘルスチェック: http://localhost:%s/health", cfg.Server.Port)
	appLogger.Info("API エンドポイント: http://localhost:%s/api/v1", cfg.Server.Port)

	if err := router.Run(":" + cfg.Server.Port); err != nil {
		appLogger.Error("サーバー起動エラー: %v", err)
		log.Fatalf("サーバー起動エラー: %v", err)
	}
}

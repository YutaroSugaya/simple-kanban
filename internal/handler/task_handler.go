package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"simple-kanban/internal/domain"
	"simple-kanban/internal/service"
	"simple-kanban/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// デバッグ用のヘルパー関数
func debugLog(format string, args ...interface{}) {
	if gin.Mode() == gin.DebugMode {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func debugRequest(c *gin.Context, data interface{}) {
	if gin.Mode() == gin.DebugMode {
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		log.Printf("[DEBUG] %s %s - Data: %s",
			c.Request.Method,
			c.Request.URL.Path,
			string(jsonData))
	}
}

func debugError(c *gin.Context, err error, context string) {
	if gin.Mode() == gin.DebugMode {
		log.Printf("[DEBUG] %s %s - Error in %s: %v",
			c.Request.Method,
			c.Request.URL.Path,
			context,
			err)
	}
}

// TaskHandler タスク関連のHTTPハンドラ
type TaskHandler struct {
	taskService service.TaskService
	validator   *validator.Validate
}

// NewTaskHandler TaskHandlerの新しいインスタンスを作成
func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		validator:   validator.New(),
	}
}

// CreateTaskRequest タスク作成リクエスト構造体
type CreateTaskRequest struct {
	ColumnID    uint    `json:"column_id" validate:"required"`
	Title       string  `json:"title" validate:"required,min=1,max=100"`
	Description string  `json:"description"`
	Order       int     `json:"order" validate:"min=1"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

// UpdateTaskRequest タスク更新リクエスト構造体
type UpdateTaskRequest struct {
	Title       *string `json:"title" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

// MoveTaskRequest タスク移動リクエスト構造体
type MoveTaskRequest struct {
	NewColumnID uint `json:"new_column_id" validate:"required"`
	NewOrder    int  `json:"new_order" validate:"min=1"`
}

// ReorderTasksRequest タスク順序変更リクエスト構造体
type ReorderTasksRequest struct {
	TaskIDs []uint `json:"task_ids" validate:"required"`
}

// CreateTask タスク作成ハンドラ
// POST /api/v1/tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	var req CreateTaskRequest

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なリクエスト形式です",
		})
		return
	}

	// バリデーション
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "バリデーションエラー",
			"details": err.Error(),
		})
		return
	}

	// 担当者IDを解析（指定されている場合）
	var assigneeID *uuid.UUID
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		parsed, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "不正な担当者IDです",
			})
			return
		}
		assigneeID = &parsed
	}

	// 期限日を解析（指定されている場合）
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "不正な期限日の形式です（YYYY-MM-DD形式で入力してください）",
			})
			return
		}
		dueDate = &parsed
	}

	// タスク作成処理
	task, err := h.taskService.CreateTask(req.ColumnID, userID, req.Title, req.Description, req.Order, assigneeID, dueDate)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	// レスポンスを構築
	response := h.buildTaskResponse(task)
	c.JSON(http.StatusCreated, gin.H{
		"task": response,
	})
}

// GetTask タスク取得ハンドラ
// GET /api/v1/tasks/:id
func (h *TaskHandler) GetTask(c *gin.Context) {
	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	// パスパラメータからタスクIDを取得
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なタスクIDです",
		})
		return
	}

	// タスクを取得
	task, err := h.taskService.GetTask(uint(taskID), userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	// レスポンスを構築
	response := h.buildTaskResponse(task)
	c.JSON(http.StatusOK, response)
}

// UpdateTask タスク更新ハンドラ
// PUT /api/v1/tasks/:id
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	debugLog("UpdateTask開始")

	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		debugError(c, err, "ユーザーID取得")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	debugLog("認証成功 - ユーザーID: %s", userID)

	// パスパラメータからタスクIDを取得
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		debugError(c, err, "タスクIDパース")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なタスクIDです",
		})
		return
	}

	debugLog("タスクID: %d", taskID)

	var req UpdateTaskRequest

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		// デバッグ用：リクエストボディを出力
		if gin.Mode() == gin.DebugMode {
			body, _ := c.GetRawData()
			log.Printf("[DEBUG] Raw request body: %s", string(body))
		}
		debugError(c, err, "JSONバインド")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "不正なリクエスト形式です",
			"details": err.Error(),
		})
		return
	}

	debugRequest(c, req)

	// バリデーション
	if err := h.validator.Struct(req); err != nil {
		debugError(c, err, "バリデーション")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "バリデーションエラー",
			"details": err.Error(),
		})
		return
	}

	// 更新データを構築
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
		debugLog("タイトル更新: %s", *req.Title)
	}
	if req.Description != nil {
		updates["description"] = *req.Description
		debugLog("説明更新: %s", *req.Description)
	}
	if req.AssigneeID != nil {
		if *req.AssigneeID == "" {
			updates["assignee_id"] = nil
			debugLog("担当者をクリア")
		} else {
			updates["assignee_id"] = *req.AssigneeID
			debugLog("担当者更新: %s", *req.AssigneeID)
		}
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			// 空文字列の場合はnullに設定
			updates["due_date"] = nil
			debugLog("期限日をクリア")
		} else {
			// 文字列をtime.Time型に変換（YYYY-MM-DD形式）
			parsed, err := time.Parse("2006-01-02", *req.DueDate)
			if err != nil {
				debugError(c, err, "日付パース")
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "不正な期限日の形式です（YYYY-MM-DD形式で入力してください）",
				})
				return
			}
			updates["due_date"] = parsed
			debugLog("期限日更新: %s", parsed.Format("2006-01-02"))
		}
	}

	debugLog("更新データ: %+v", updates)

	// タスク更新処理
	task, err := h.taskService.UpdateTask(uint(taskID), userID, updates)
	if err != nil {
		debugError(c, err, "タスク更新")
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	debugLog("タスク更新成功: %+v", task)

	// レスポンスを構築
	response := h.buildTaskResponse(task)
	c.JSON(http.StatusOK, response)
}

// DeleteTask タスク削除ハンドラ
// DELETE /api/v1/tasks/:id
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	// パスパラメータからタスクIDを取得
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なタスクIDです",
		})
		return
	}

	// タスク削除処理
	if err := h.taskService.DeleteTask(uint(taskID), userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// MoveTask タスク移動ハンドラ
// PUT /api/v1/tasks/:id/move
func (h *TaskHandler) MoveTask(c *gin.Context) {
	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	// パスパラメータからタスクIDを取得
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なタスクIDです",
		})
		return
	}

	var req MoveTaskRequest

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なリクエスト形式です",
		})
		return
	}

	// バリデーション
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "バリデーションエラー",
			"details": err.Error(),
		})
		return
	}

	// タスク移動処理
	if err := h.taskService.MoveTask(uint(taskID), req.NewColumnID, req.NewOrder, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "タスクが正常に移動されました",
	})
}

// ReorderTasks タスク順序変更ハンドラ
// PUT /api/v1/columns/:columnId/tasks/reorder
func (h *TaskHandler) ReorderTasks(c *gin.Context) {
	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	// パスパラメータからカラムIDを取得
	columnIDStr := c.Param("columnId")
	columnID, err := strconv.ParseUint(columnIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なカラムIDです",
		})
		return
	}

	var req ReorderTasksRequest

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なリクエスト形式です",
		})
		return
	}

	// バリデーション
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "バリデーションエラー",
			"details": err.Error(),
		})
		return
	}

	// タスク順序変更処理
	if err := h.taskService.ReorderTasks(uint(columnID), req.TaskIDs, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "タスクの順序が正常に更新されました",
	})
}

// buildTaskResponse タスクレスポンスを構築するヘルパー関数
func (h *TaskHandler) buildTaskResponse(task *domain.Task) TaskResponse {
	response := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Order:       task.Order,
		DueDate:     task.DueDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}

	// 担当者情報が存在する場合は追加
	if task.AssigneeID != nil {
		assigneeIDStr := task.AssigneeID.String()
		response.AssigneeID = &assigneeIDStr
	}
	if task.Assignee != nil {
		response.Assignee = &UserResponse{
			ID:    task.Assignee.ID.String(),
			Email: task.Assignee.Email,
		}
	}

	return response
}

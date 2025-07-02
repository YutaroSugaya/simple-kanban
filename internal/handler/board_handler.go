package handler

import (
	"net/http"
	"strconv"
	"time"

	"simple-kanban/internal/service"
	"simple-kanban/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// BoardHandler ボード関連のHTTPハンドラ
type BoardHandler struct {
	boardService service.BoardService
	validator    *validator.Validate
}

// NewBoardHandler BoardHandlerの新しいインスタンスを作成
func NewBoardHandler(boardService service.BoardService) *BoardHandler {
	return &BoardHandler{
		boardService: boardService,
		validator:    validator.New(),
	}
}

// CreateBoardRequest ボード作成リクエスト構造体
type CreateBoardRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// UpdateBoardRequest ボード更新リクエスト構造体
type UpdateBoardRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// BoardResponse ボード情報レスポンス構造体
type BoardResponse struct {
	ID        uint             `json:"id"`
	Name      string           `json:"name"`
	OwnerID   string           `json:"owner_id"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Columns   []ColumnResponse `json:"columns,omitempty"`
}

// ColumnResponse カラム情報レスポンス構造体
type ColumnResponse struct {
	ID    uint           `json:"id"`
	Title string         `json:"title"`
	Order int            `json:"order"`
	Tasks []TaskResponse `json:"tasks,omitempty"`
}

// TaskResponse タスク情報レスポンス構造体
type TaskResponse struct {
	ID          uint          `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Order       int           `json:"order"`
	AssigneeID  *string       `json:"assignee_id"`
	Assignee    *UserResponse `json:"assignee,omitempty"`
	DueDate     *time.Time    `json:"due_date"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// CreateBoard ボード作成ハンドラ
// POST /api/v1/boards
func (h *BoardHandler) CreateBoard(c *gin.Context) {
	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	var req CreateBoardRequest

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

	// ボード作成処理
	board, err := h.boardService.CreateBoard(userID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// レスポンスを返す
	response := BoardResponse{
		ID:        board.ID,
		Name:      board.Name,
		OwnerID:   board.OwnerID.String(),
		CreatedAt: board.CreatedAt,
		UpdatedAt: board.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"board": response,
	})
}

// GetUserBoards ユーザーのボード一覧取得ハンドラ
// GET /api/v1/boards
func (h *BoardHandler) GetUserBoards(c *gin.Context) {
	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	// ユーザーのボード一覧を取得
	boards, err := h.boardService.GetUserBoards(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// レスポンスを構築
	var response []BoardResponse
	for _, board := range boards {
		response = append(response, BoardResponse{
			ID:        board.ID,
			Name:      board.Name,
			OwnerID:   board.OwnerID.String(),
			CreatedAt: board.CreatedAt,
			UpdatedAt: board.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"boards": response,
	})
}

// GetBoardWithColumns ボードとカラム情報取得ハンドラ
// GET /api/v1/boards/:id/columns
func (h *BoardHandler) GetBoardWithColumns(c *gin.Context) {
	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	// パスパラメータからボードIDを取得
	boardIDStr := c.Param("id")
	boardID, err := strconv.ParseUint(boardIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なボードIDです",
		})
		return
	}

	// ボードをカラム情報付きで取得
	board, err := h.boardService.GetBoardWithColumns(uint(boardID), userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	// レスポンスを構築
	var columns []ColumnResponse
	for _, column := range board.Columns {
		var tasks []TaskResponse
		for _, task := range column.Tasks {
			taskResponse := TaskResponse{
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
				taskResponse.AssigneeID = &assigneeIDStr
			}
			if task.Assignee != nil {
				taskResponse.Assignee = &UserResponse{
					ID:    task.Assignee.ID.String(),
					Email: task.Assignee.Email,
				}
			}

			tasks = append(tasks, taskResponse)
		}

		columns = append(columns, ColumnResponse{
			ID:    column.ID,
			Title: column.Title,
			Order: column.Order,
			Tasks: tasks,
		})
	}

	response := BoardResponse{
		ID:        board.ID,
		Name:      board.Name,
		OwnerID:   board.OwnerID.String(),
		CreatedAt: board.CreatedAt,
		UpdatedAt: board.UpdatedAt,
		Columns:   columns,
	}

	c.JSON(http.StatusOK, gin.H{
		"board": response,
	})
}

// UpdateBoard ボード更新ハンドラ
// PUT /api/v1/boards/:id
func (h *BoardHandler) UpdateBoard(c *gin.Context) {
	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	// パスパラメータからボードIDを取得
	boardIDStr := c.Param("id")
	boardID, err := strconv.ParseUint(boardIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なボードIDです",
		})
		return
	}

	var req UpdateBoardRequest

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

	// ボード更新処理
	updates := map[string]interface{}{
		"name": req.Name,
	}
	board, err := h.boardService.UpdateBoard(uint(boardID), userID, updates)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	// レスポンスを返す
	response := BoardResponse{
		ID:        board.ID,
		Name:      board.Name,
		OwnerID:   board.OwnerID.String(),
		CreatedAt: board.CreatedAt,
		UpdatedAt: board.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"board": response,
	})
}

// DeleteBoard ボード削除ハンドラ
// DELETE /api/v1/boards/:id
func (h *BoardHandler) DeleteBoard(c *gin.Context) {
	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	// パスパラメータからボードIDを取得
	boardIDStr := c.Param("id")
	boardID, err := strconv.ParseUint(boardIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なボードIDです",
		})
		return
	}

	// ボード削除処理
	if err := h.boardService.DeleteBoard(uint(boardID), userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

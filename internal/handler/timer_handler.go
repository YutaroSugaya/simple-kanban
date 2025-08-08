package handler

import (
	"net/http"
	"simple-kanban/internal/service"
	"simple-kanban/pkg/middleware"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TimerHandler タイマー関連のHTTPハンドラー
type TimerHandler struct {
	timerService service.TimerService
}

// NewTimerHandler タイマーハンドラーのコンストラクタ
func NewTimerHandler(timerService service.TimerService) *TimerHandler {
	return &TimerHandler{
		timerService: timerService,
	}
}

// StartTimer タイマーを開始
// @Summary タイマー開始
// @Description タスクのタイマーを開始します
// @Tags timer
// @Accept json
// @Produce json
// @Param request body StartTimerRequest true "タイマー開始リクエスト"
// @Success 201 {object} domain.TimerSession
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/timer/start [post]
func (h *TimerHandler) StartTimer(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	var request StartTimerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "リクエストデータが無効です"})
		return
	}

	session, err := h.timerService.StartTimer(userID, request.TaskID, request.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// StopTimer タイマーを停止
// @Summary タイマー停止
// @Description タイマーを停止します
// @Tags timer
// @Accept json
// @Produce json
// @Param id path int true "セッションID"
// @Success 200 {object} domain.TimerSession
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/timer/{id}/stop [put]
func (h *TimerHandler) StopTimer(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	sessionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "無効なセッションIDです"})
		return
	}

	session, err := h.timerService.StopTimer(userID, uint(sessionID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetActiveTimer アクティブなタイマーを取得
// @Summary アクティブタイマー取得
// @Description ユーザーのアクティブなタイマーを取得します
// @Tags timer
// @Accept json
// @Produce json
// @Success 200 {object} domain.TimerSession
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/timer/active [get]
func (h *TimerHandler) GetActiveTimer(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	session, err := h.timerService.GetActiveTimer(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "アクティブなタイマーが見つかりません"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetTimerHistory タイマー履歴を取得
// @Summary タイマー履歴取得
// @Description ユーザーのタイマー履歴を取得します
// @Tags timer
// @Accept json
// @Produce json
// @Success 200 {array} domain.TimerSession
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/timer/history [get]
func (h *TimerHandler) GetTimerHistory(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	sessions, err := h.timerService.GetTimerHistory(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// GetTimersByTask タスクのタイマー履歴を取得
// @Summary タスク別タイマー履歴取得
// @Description 指定タスクのタイマー履歴を取得します
// @Tags timer
// @Accept json
// @Produce json
// @Param taskId path int true "タスクID"
// @Success 200 {array} domain.TimerSession
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/timer/tasks/{taskId} [get]
func (h *TimerHandler) GetTimersByTask(c *gin.Context) {
	_, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "無効なタスクIDです"})
		return
	}

	sessions, err := h.timerService.GetTimersByTask(uint(taskID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// StartTimerRequest タイマー開始のリクエスト
type StartTimerRequest struct {
	TaskID   uint `json:"task_id" binding:"required"`
	Duration int  `json:"duration" binding:"required"` // 秒
}

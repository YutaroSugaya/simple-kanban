package handler

import (
	"net/http"
	"simple-kanban/internal/domain"
	"simple-kanban/internal/service"
	"simple-kanban/pkg/logger"
	"simple-kanban/pkg/middleware"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CalendarHandler カレンダー関連のHTTPハンドラー
type CalendarHandler struct {
	calendarService service.CalendarService
	taskService     service.TaskService
	logger          *logger.Logger
}

// NewCalendarHandler カレンダーハンドラーのコンストラクタ
func NewCalendarHandler(calendarService service.CalendarService, taskService service.TaskService, logger *logger.Logger) *CalendarHandler {
	return &CalendarHandler{
		calendarService: calendarService,
		taskService:     taskService,
		logger:          logger,
	}
}

// GetCalendarSettings カレンダー設定を取得
// @Summary カレンダー設定取得
// @Description ユーザーのカレンダー設定を取得します
// @Tags calendar
// @Accept json
// @Produce json
// @Success 200 {object} domain.CalendarSettings
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/calendar/settings [get]
func (h *CalendarHandler) GetCalendarSettings(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	settings, err := h.calendarService.GetCalendarSettings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateCalendarSettings カレンダー設定を更新
// @Summary カレンダー設定更新
// @Description ユーザーのカレンダー設定を更新します
// @Tags calendar
// @Accept json
// @Produce json
// @Param settings body domain.CalendarSettings true "カレンダー設定"
// @Success 200 {object} domain.CalendarSettings
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/calendar/settings [put]
func (h *CalendarHandler) UpdateCalendarSettings(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	var settings domain.CalendarSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "リクエストデータが無効です"})
		return
	}

	if err := h.calendarService.UpdateCalendarSettings(userID, &settings); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// 更新後の設定を取得して返す
	updatedSettings, err := h.calendarService.GetCalendarSettings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedSettings)
}

// GetEvents カレンダーイベントを取得
// @Summary カレンダーイベント取得
// @Description 指定期間のカレンダーイベントを取得します
// @Tags calendar
// @Accept json
// @Produce json
// @Param start query string true "開始日時 (RFC3339形式)"
// @Param end query string true "終了日時 (RFC3339形式)"
// @Success 200 {array} domain.CalendarEvent
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/calendar/events [get]
func (h *CalendarHandler) GetEvents(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "start と end パラメータが必要です"})
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "start パラメータの形式が無効です"})
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "end パラメータの形式が無効です"})
		return
	}

	events, err := h.calendarService.GetEventsByDateRange(userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

// CreateEvent カレンダーイベントを作成
// @Summary カレンダーイベント作成
// @Description 新しいカレンダーイベントを作成します
// @Tags calendar
// @Accept json
// @Produce json
// @Param event body domain.CalendarEvent true "カレンダーイベント"
// @Success 201 {object} domain.CalendarEvent
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/calendar/events [post]
func (h *CalendarHandler) CreateEvent(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	var event domain.CalendarEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "リクエストデータが無効です"})
		return
	}

	if err := h.calendarService.CreateEvent(userID, &event); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// UpdateEvent カレンダーイベントを更新
// @Summary カレンダーイベント更新
// @Description カレンダーイベントを更新します
// @Tags calendar
// @Accept json
// @Produce json
// @Param id path int true "イベントID"
// @Param event body domain.CalendarEvent true "カレンダーイベント"
// @Success 200 {object} domain.CalendarEvent
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/calendar/events/{id} [put]
func (h *CalendarHandler) UpdateEvent(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	eventID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "無効なイベントIDです"})
		return
	}

	var event domain.CalendarEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "リクエストデータが無効です"})
		return
	}

	if err := h.calendarService.UpdateEvent(userID, uint(eventID), &event); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

// DeleteEvent カレンダーイベントを削除
// @Summary カレンダーイベント削除
// @Description カレンダーイベントを削除します
// @Tags calendar
// @Accept json
// @Produce json
// @Param id path int true "イベントID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/calendar/events/{id} [delete]
func (h *CalendarHandler) DeleteEvent(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	eventID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "無効なイベントIDです"})
		return
	}

	if err := h.calendarService.DeleteEvent(userID, uint(eventID)); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// CreateTaskEvent タスクからカレンダーイベントを作成
// @Summary タスクからカレンダーイベント作成
// @Description タスクを基にカレンダーイベントを作成します
// @Tags calendar
// @Accept json
// @Produce json
// @Param taskId path int true "タスクID"
// @Param event body CreateTaskEventRequest true "イベント作成リクエスト"
// @Success 201 {string} string "Created"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/calendar/tasks/{taskId}/events [post]
func (h *CalendarHandler) CreateTaskEvent(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		h.logger.Error("CreateTaskEvent: 認証情報取得エラー - %v", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	taskID, err := strconv.ParseUint(c.Param("taskId"), 10, 32)
	if err != nil {
		h.logger.Error("CreateTaskEvent: 無効なタスクID - %s, エラー: %v", c.Param("taskId"), err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "無効なタスクIDです"})
		return
	}

	var request CreateTaskEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("CreateTaskEvent: リクエストデータバインドエラー - %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "リクエストデータが無効です"})
		return
	}

	h.logger.Info("CreateTaskEvent: リクエスト処理開始 - UserID: %s, TaskID: %d, Start: %v, End: %v",
		userID, taskID, request.Start, request.End)

	// タスクを取得
	task, err := h.taskService.GetTask(uint(taskID), userID)
	if err != nil {
		h.logger.Error("CreateTaskEvent: タスク取得エラー - TaskID: %d, UserID: %s, エラー: %v", taskID, userID, err)
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "タスクが見つかりません"})
		return
	}

	h.logger.Info("CreateTaskEvent: タスク取得成功 - TaskID: %d, Title: %s", task.ID, task.Title)

	if err := h.calendarService.CreateEventFromTask(userID, task, request.Start, request.End); err != nil {
		h.logger.Error("CreateTaskEvent: カレンダーイベント作成エラー - UserID: %s, TaskID: %d, エラー: %v",
			userID, taskID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info("CreateTaskEvent: カレンダーイベント作成成功 - UserID: %s, TaskID: %d", userID, taskID)
	c.JSON(http.StatusCreated, gin.H{"message": "カレンダーイベントが正常に作成されました"})
}

// CreateTaskEventRequest タスクからカレンダーイベント作成のリクエスト
type CreateTaskEventRequest struct {
	Start time.Time `json:"start" binding:"required"`
	End   time.Time `json:"end" binding:"required"`
}

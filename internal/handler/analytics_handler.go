package handler

import (
	"net/http"
	"simple-kanban/pkg/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticsHandler 分析関連のHTTPハンドラー
type AnalyticsHandler struct {
	// 必要なサービスやリポジトリを追加予定
}

// NewAnalyticsHandler 分析ハンドラーのコンストラクタ
func NewAnalyticsHandler() *AnalyticsHandler {
	return &AnalyticsHandler{}
}

// GetTaskCompletionStats タスク完了統計を取得
// @Summary タスク完了統計取得
// @Description 指定年のタスク完了統計を取得します
// @Tags analytics
// @Accept json
// @Produce json
// @Param year query string false "年 (YYYY形式)"
// @Success 200 {object} TaskCompletionStats
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/analytics/task-completion [get]
func (h *AnalyticsHandler) GetTaskCompletionStats(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "認証情報が取得できません"})
		return
	}

	// 年の指定（デフォルトは現在年）
	year := time.Now().Year()
	if yearParam := c.Query("year"); yearParam != "" {
		if parsedYear, err := time.Parse("2006", yearParam); err == nil {
			year = parsedYear.Year()
		}
	}

	// 年の開始日と終了日を計算
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	// 日毎の完了タスク数を取得
	stats, err := h.getTaskCompletionByDate(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getTaskCompletionByDate 指定期間の日毎タスク完了数を取得
func (h *AnalyticsHandler) getTaskCompletionByDate(userID interface{}, startDate, endDate time.Time) (interface{}, error) {
	// TODO: 実際のデータベースからタスク完了統計を取得する実装
	// 現在はダミーデータを返す
	stats := map[string]interface{}{
		"year":        startDate.Year(),
		"total_tasks": 150,
		"daily_stats": map[string]int{
			"2024-01-01": 3,
			"2024-01-02": 5,
			"2024-01-03": 2,
			// 他の日のデータ...
		},
	}
	return stats, nil
}

// TaskCompletionStats タスク完了統計のレスポンス
type TaskCompletionStats struct {
	Year       int            `json:"year"`
	TotalTasks int            `json:"total_tasks"`
	DailyStats map[string]int `json:"daily_stats"`
}

package handler

import (
	"net/http"
	"simple-kanban/pkg/middleware"
	"time"

	"simple-kanban/internal/repository"

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

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	stats, err := h.getTaskCompletionByDate(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getTaskCompletionByDate 指定期間の日毎タスク完了数を取得
func (h *AnalyticsHandler) getTaskCompletionByDate(userID interface{}, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// DB から tasks と boards をJOINして、対象ユーザーのボード配下の完了タスクを日付ごとに集計
	db := repository.GetDB()
	type row struct {
		Date  time.Time
		Count int
	}
	var rows []row
	q := db.Table("tasks t").
		Select("DATE(t.updated_at) as date, COUNT(*) as count").
		Joins("JOIN columns c ON c.id = t.column_id").
		Joins("JOIN boards b ON b.id = c.board_id").
		Where("b.owner_id = ? AND t.is_completed = ? AND t.updated_at >= ? AND t.updated_at < ?", userID, true, startDate, endDate).
		Group("DATE(t.updated_at)").
		Order("DATE(t.updated_at)")
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}

	// 返却形式に整形
	var result []map[string]interface{}
	for _, r := range rows {
		result = append(result, map[string]interface{}{
			"date":  r.Date.Format("2006-01-02"),
			"count": r.Count,
		})
	}
	return result, nil
}

// TaskCompletionStats タスク完了統計のレスポンス（未使用）
type TaskCompletionStats struct {
	Year       int            `json:"year"`
	TotalTasks int            `json:"total_tasks"`
	DailyStats map[string]int `json:"daily_stats"`
}

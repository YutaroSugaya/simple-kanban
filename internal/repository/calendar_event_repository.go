package repository

import (
	"simple-kanban/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CalendarEventRepository カレンダーイベントのリポジトリインターフェース
type CalendarEventRepository interface {
	Create(event *domain.CalendarEvent) error
	GetByID(id uint) (*domain.CalendarEvent, error)
	GetByUserID(userID uuid.UUID) ([]*domain.CalendarEvent, error)
	GetByUserIDAndDateRange(userID uuid.UUID, start, end time.Time) ([]*domain.CalendarEvent, error)
	GetByTaskID(taskID uint) (*domain.CalendarEvent, error)
	Update(event *domain.CalendarEvent) error
	Delete(id uint) error
}

// calendarEventRepository カレンダーイベントリポジトリの実装
type calendarEventRepository struct {
	db *gorm.DB
}

// NewCalendarEventRepository カレンダーイベントリポジトリのコンストラクタ
func NewCalendarEventRepository(db *gorm.DB) CalendarEventRepository {
	return &calendarEventRepository{db: db}
}

// Create カレンダーイベントを作成します
func (r *calendarEventRepository) Create(event *domain.CalendarEvent) error {
	return r.db.Create(event).Error
}

// GetByID IDでカレンダーイベントを取得します
func (r *calendarEventRepository) GetByID(id uint) (*domain.CalendarEvent, error) {
	var event domain.CalendarEvent
	err := r.db.Preload("User").Preload("Task").First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GetByUserID ユーザーIDでカレンダーイベントを取得します
func (r *calendarEventRepository) GetByUserID(userID uuid.UUID) ([]*domain.CalendarEvent, error) {
	var events []*domain.CalendarEvent
	err := r.db.Where("user_id = ?", userID).
		Preload("User").Preload("Task").
		Order("\"start\" ASC").
		Find(&events).Error
	return events, err
}

// GetByUserIDAndDateRange ユーザーIDと日付範囲でカレンダーイベントを取得します
func (r *calendarEventRepository) GetByUserIDAndDateRange(userID uuid.UUID, start, end time.Time) ([]*domain.CalendarEvent, error) {
	var events []*domain.CalendarEvent
	err := r.db.Where("user_id = ? AND \"start\" < ? AND \"end\" > ?", userID, end, start).
		Preload("User").Preload("Task").
		Order("\"start\" ASC").
		Find(&events).Error
	return events, err
}

// GetByTaskID タスクIDでカレンダーイベントを取得します
func (r *calendarEventRepository) GetByTaskID(taskID uint) (*domain.CalendarEvent, error) {
	var event domain.CalendarEvent
	err := r.db.Where("task_id = ?", taskID).
		Preload("User").Preload("Task").
		First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// Update カレンダーイベントを更新します
func (r *calendarEventRepository) Update(event *domain.CalendarEvent) error {
	return r.db.Save(event).Error
}

// Delete カレンダーイベントを削除します
func (r *calendarEventRepository) Delete(id uint) error {
	return r.db.Delete(&domain.CalendarEvent{}, id).Error
}

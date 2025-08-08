package repository

import (
	"simple-kanban/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CalendarSettingsRepository カレンダー設定のリポジトリインターフェース
type CalendarSettingsRepository interface {
	Create(settings *domain.CalendarSettings) error
	GetByUserID(userID uuid.UUID) (*domain.CalendarSettings, error)
	Update(settings *domain.CalendarSettings) error
	Delete(id uint) error
}

// calendarSettingsRepository カレンダー設定リポジトリの実装
type calendarSettingsRepository struct {
	db *gorm.DB
}

// NewCalendarSettingsRepository カレンダー設定リポジトリのコンストラクタ
func NewCalendarSettingsRepository(db *gorm.DB) CalendarSettingsRepository {
	return &calendarSettingsRepository{db: db}
}

// Create カレンダー設定を作成します
func (r *calendarSettingsRepository) Create(settings *domain.CalendarSettings) error {
	return r.db.Create(settings).Error
}

// GetByUserID ユーザーIDでカレンダー設定を取得します
func (r *calendarSettingsRepository) GetByUserID(userID uuid.UUID) (*domain.CalendarSettings, error) {
	var settings domain.CalendarSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// Update カレンダー設定を更新します
func (r *calendarSettingsRepository) Update(settings *domain.CalendarSettings) error {
	return r.db.Save(settings).Error
}

// Delete カレンダー設定を削除します
func (r *calendarSettingsRepository) Delete(id uint) error {
	return r.db.Delete(&domain.CalendarSettings{}, id).Error
}

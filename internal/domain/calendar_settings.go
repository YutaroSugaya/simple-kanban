package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CalendarSettings ユーザーのカレンダー表示設定を表すエンティティ
type CalendarSettings struct {
	ID               uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID           uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	WeekdayStartTime string         `json:"weekday_start_time" gorm:"not null;default:'09:00'"` // 平日開始時刻
	WeekdayEndTime   string         `json:"weekday_end_time" gorm:"not null;default:'18:00'"`   // 平日終了時刻
	WeekendStartTime string         `json:"weekend_start_time" gorm:"not null;default:'10:00'"` // 土日開始時刻
	WeekendEndTime   string         `json:"weekend_end_time" gorm:"not null;default:'16:00'"`   // 土日終了時刻
	TimeSlotDuration int            `json:"time_slot_duration" gorm:"not null;default:10"`      // 時間スロット（分）
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`

	// リレーション：この設定の所有者
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName テーブル名を明示的に指定
func (CalendarSettings) TableName() string {
	return "calendar_settings"
}

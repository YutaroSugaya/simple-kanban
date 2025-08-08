package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CalendarEvent カレンダー上のイベントを表すエンティティ
type CalendarEvent struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	TaskID      *uint          `json:"task_id,omitempty" gorm:"index"` // タスクベースの場合のタスクID
	Title       string         `json:"title" gorm:"not null" validate:"required"`
	Start       time.Time      `json:"start" gorm:"not null"`
	End         time.Time      `json:"end" gorm:"not null"`
	Color       string         `json:"color" gorm:"default:'#3B82F6'"`              // イベントの色
	IsTaskBased bool           `json:"is_task_based" gorm:"not null;default:false"` // タスクベースのイベントかどうか
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// リレーション：このイベントの所有者
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// リレーション：このイベントが関連するタスク（任意）
	Task *Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`
}

// TableName テーブル名を明示的に指定
func (CalendarEvent) TableName() string {
	return "calendar_events"
}

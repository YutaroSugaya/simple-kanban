package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TimerSession タスクのタイマーセッションを表すエンティティ
type TimerSession struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID    uint           `json:"task_id" gorm:"not null;index"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	StartTime time.Time      `json:"start_time" gorm:"not null"`
	EndTime   *time.Time     `json:"end_time,omitempty" gorm:"default:null"`
	Duration  int            `json:"duration" gorm:"not null;default:0"`     // 継続時間（秒）
	IsActive  bool           `json:"is_active" gorm:"not null;default:true"` // アクティブ状態
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// リレーション：このセッションが属するタスク
	Task Task `json:"task,omitempty" gorm:"foreignKey:TaskID"`

	// リレーション：このセッションの実行者
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName テーブル名を明示的に指定
func (TimerSession) TableName() string {
	return "timer_sessions"
}

package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task Kanbanボードのタスクを表すエンティティ
// カラムに属し、ユーザーが担当できます
type Task struct {
	ID          uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	ColumnID    uint       `json:"column_id" gorm:"not null;index"`
	Title       string     `json:"title" gorm:"not null" validate:"required,min=1,max=100"`
	Description string     `json:"description" gorm:"type:text"`
	Order       int        `json:"order" gorm:"not null;default:0"`              // タスクの表示順序
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty" gorm:"type:uuid;index"` // 担当者（任意）
	DueDate     *time.Time `json:"due_date,omitempty"`                           // 期限（任意）

	// 新機能用フィールド
	EstimatedTime  *int       `json:"estimated_time,omitempty" gorm:"default:null"`  // 目標時間（分）
	ActualTime     *int       `json:"actual_time,omitempty" gorm:"default:null"`     // 実際にかかった時間（分）
	IsCompleted    bool       `json:"is_completed" gorm:"default:false"`             // 完了状態
	ScheduledStart *time.Time `json:"scheduled_start,omitempty" gorm:"default:null"` // スケジュール開始時刻
	ScheduledEnd   *time.Time `json:"scheduled_end,omitempty" gorm:"default:null"`   // スケジュール終了時刻
	CalendarDate   *time.Time `json:"calendar_date,omitempty" gorm:"default:null"`   // カレンダー配置日

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // ソフトデリート対応

	// リレーション：このタスクが属するカラム
	Column Column `json:"column,omitempty" gorm:"foreignKey:ColumnID"`

	// リレーション：このタスクの担当者（任意）
	Assignee *User `json:"assignee,omitempty" gorm:"foreignKey:AssigneeID"`
}

// TableName テーブル名を明示的に指定
func (Task) TableName() string {
	return "tasks"
}

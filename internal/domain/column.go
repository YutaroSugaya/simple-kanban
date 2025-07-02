package domain

import (
	"time"

	"gorm.io/gorm"
)

// Column Kanbanボードのカラム（列）を表すエンティティ
// ボードに属し、複数のタスクを持ちます
type Column struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	BoardID   uint           `json:"board_id" gorm:"not null;index"`
	Title     string         `json:"title" gorm:"not null" validate:"required,min=1,max=50"`
	Order     int            `json:"order" gorm:"not null;default:0"` // カラムの表示順序
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // ソフトデリート対応

	// リレーション：このカラムが属するボード
	Board Board `json:"board,omitempty" gorm:"foreignKey:BoardID"`

	// リレーション：このカラムが持つタスク一覧（order順でソート）
	Tasks []Task `json:"tasks,omitempty" gorm:"foreignKey:ColumnID;constraint:OnDelete:CASCADE"`
}

// TableName テーブル名を明示的に指定
func (Column) TableName() string {
	return "columns"
}

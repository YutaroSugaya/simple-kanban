package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Board Kanbanボードを表すエンティティ
// ユーザーが所有し、複数のカラムを持ちます
type Board struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string         `json:"name" gorm:"not null" validate:"required,min=1,max=100"`
	OwnerID   uuid.UUID      `json:"owner_id" gorm:"type:uuid;not null;index"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // ソフトデリート対応

	// リレーション：このボードの所有者
	Owner User `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	
	// リレーション：このボードが持つカラム一覧（order順でソート）
	Columns []Column `json:"columns,omitempty" gorm:"foreignKey:BoardID;constraint:OnDelete:CASCADE"`
}

// TableName テーブル名を明示的に指定
func (Board) TableName() string {
	return "boards"
} 
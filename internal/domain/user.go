package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User ユーザー情報を表すエンティティ
// 認証とボードの所有者情報を管理します
type User struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email        string         `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	PasswordHash string         `json:"-" gorm:"not null"` // JSONでは出力しない（セキュリティ）
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"` // ソフトデリート対応

	// リレーション：ユーザーが所有するボード一覧
	Boards []Board `json:"boards,omitempty" gorm:"foreignKey:OwnerID"`
}

// TableName テーブル名を明示的に指定
func (User) TableName() string {
	return "users"
}

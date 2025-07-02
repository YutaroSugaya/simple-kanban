package repository

import (
	"simple-kanban/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository ユーザーのデータアクセスを管理するインターフェース
type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id uuid.UUID) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id uuid.UUID) error
	List(limit, offset int) ([]domain.User, error)
}

// userRepository UserRepositoryの実装
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository UserRepositoryの新しいインスタンスを作成
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 新しいユーザーを作成します
func (r *userRepository) Create(user *domain.User) error {
	// UUIDを生成（データベースで自動生成されない場合）
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	result := r.db.Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByID IDでユーザーを取得します
func (r *userRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // ユーザーが見つからない場合はnilを返す
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByEmail メールアドレスでユーザーを取得します（認証で使用）
func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // ユーザーが見つからない場合はnilを返す
		}
		return nil, result.Error
	}
	return &user, nil
}

// Update ユーザー情報を更新します
func (r *userRepository) Update(user *domain.User) error {
	result := r.db.Save(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete ユーザーを削除します（ソフトデリート）
func (r *userRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&domain.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// List ユーザー一覧を取得します（ページネーション対応）
func (r *userRepository) List(limit, offset int) ([]domain.User, error) {
	var users []domain.User
	result := r.db.Limit(limit).Offset(offset).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

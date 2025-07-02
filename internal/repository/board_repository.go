package repository

import (
	"simple-kanban/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BoardRepository ボードのデータアクセスを管理するインターフェース
type BoardRepository interface {
	Create(board *domain.Board) error
	GetByID(id uint) (*domain.Board, error)
	GetByOwnerID(ownerID uuid.UUID) ([]domain.Board, error)
	GetByIDWithColumns(id uint) (*domain.Board, error)
	Update(board *domain.Board) error
	Delete(id uint) error
	List(limit, offset int) ([]domain.Board, error)
}

// boardRepository BoardRepositoryの実装
type boardRepository struct {
	db *gorm.DB
}

// NewBoardRepository BoardRepositoryの新しいインスタンスを作成
func NewBoardRepository(db *gorm.DB) BoardRepository {
	return &boardRepository{db: db}
}

// Create 新しいボードを作成します
func (r *boardRepository) Create(board *domain.Board) error {
	result := r.db.Create(board)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByID IDでボードを取得します
func (r *boardRepository) GetByID(id uint) (*domain.Board, error) {
	var board domain.Board
	result := r.db.Where("id = ?", id).First(&board)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // ボードが見つからない場合はnilを返す
		}
		return nil, result.Error
	}
	return &board, nil
}

// GetByOwnerID 所有者IDでボード一覧を取得します
func (r *boardRepository) GetByOwnerID(ownerID uuid.UUID) ([]domain.Board, error) {
	var boards []domain.Board
	result := r.db.Where("owner_id = ?", ownerID).Find(&boards)
	if result.Error != nil {
		return nil, result.Error
	}
	return boards, nil
}

// GetByIDWithColumns IDでボードを取得し、カラム情報も含めます
func (r *boardRepository) GetByIDWithColumns(id uint) (*domain.Board, error) {
	var board domain.Board
	result := r.db.Preload("Columns", func(db *gorm.DB) *gorm.DB {
		return db.Order("columns.\"order\" ASC")
	}).Preload("Columns.Tasks", func(db *gorm.DB) *gorm.DB {
		return db.Order("tasks.\"order\" ASC")
	}).Preload("Columns.Tasks.Assignee").Where("id = ?", id).First(&board)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // ボードが見つからない場合はnilを返す
		}
		return nil, result.Error
	}
	return &board, nil
}

// Update ボード情報を更新します
func (r *boardRepository) Update(board *domain.Board) error {
	result := r.db.Save(board)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete ボードを削除します（ソフトデリート）
func (r *boardRepository) Delete(id uint) error {
	result := r.db.Delete(&domain.Board{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// List ボード一覧を取得します（ページネーション対応）
func (r *boardRepository) List(limit, offset int) ([]domain.Board, error) {
	var boards []domain.Board
	result := r.db.Preload("Owner").Limit(limit).Offset(offset).Find(&boards)
	if result.Error != nil {
		return nil, result.Error
	}
	return boards, nil
}

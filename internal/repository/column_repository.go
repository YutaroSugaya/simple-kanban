package repository

import (
	"simple-kanban/internal/domain"

	"gorm.io/gorm"
)

// ColumnRepository カラムのデータアクセスを管理するインターフェース
type ColumnRepository interface {
	Create(column *domain.Column) error
	GetByID(id uint) (*domain.Column, error)
	GetByBoardID(boardID uint) ([]domain.Column, error)
	Update(column *domain.Column) error
	Delete(id uint) error
	UpdateOrder(id uint, newOrder int) error
	ReorderColumns(boardID uint, columnIDs []uint) error
}

// columnRepository ColumnRepositoryの実装
type columnRepository struct {
	db *gorm.DB
}

// NewColumnRepository ColumnRepositoryの新しいインスタンスを作成
func NewColumnRepository(db *gorm.DB) ColumnRepository {
	return &columnRepository{db: db}
}

// Create 新しいカラムを作成します
func (r *columnRepository) Create(column *domain.Column) error {
	// 新しいカラムを作成する際は、そのボードの最後の順序番号を取得して+1する
	if column.Order == 0 {
		var maxOrder int
		r.db.Model(&domain.Column{}).Where("board_id = ?", column.BoardID).Select("COALESCE(MAX(\"order\"), 0)").Scan(&maxOrder)
		column.Order = maxOrder + 1
	}

	result := r.db.Create(column)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByID IDでカラムを取得します
func (r *columnRepository) GetByID(id uint) (*domain.Column, error) {
	var column domain.Column
	result := r.db.Preload("Board").Preload("Tasks", func(db *gorm.DB) *gorm.DB {
		return db.Order("tasks.\"order\" ASC")
	}).Preload("Tasks.Assignee").Where("id = ?", id).First(&column)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // カラムが見つからない場合はnilを返す
		}
		return nil, result.Error
	}
	return &column, nil
}

// GetByBoardID ボードIDでカラム一覧を取得します（順序順）
func (r *columnRepository) GetByBoardID(boardID uint) ([]domain.Column, error) {
	var columns []domain.Column
	result := r.db.Preload("Tasks", func(db *gorm.DB) *gorm.DB {
		return db.Order("tasks.\"order\" ASC")
	}).Preload("Tasks.Assignee").Where("board_id = ?", boardID).Order("\"order\" ASC").Find(&columns)

	if result.Error != nil {
		return nil, result.Error
	}
	return columns, nil
}

// Update カラム情報を更新します
func (r *columnRepository) Update(column *domain.Column) error {
	result := r.db.Save(column)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete カラムを削除します（ソフトデリート）
func (r *columnRepository) Delete(id uint) error {
	// カラムを削除する前に、同じボード内の他のカラムの順序を調整
	var column domain.Column
	if err := r.db.First(&column, id).Error; err != nil {
		return err
	}

	// トランザクション内で削除と順序調整を実行
	return r.db.Transaction(func(tx *gorm.DB) error {
		// カラムを削除（カスケード削除でタスクも削除される）
		if err := tx.Delete(&domain.Column{}, id).Error; err != nil {
			return err
		}

		// 削除されたカラムより後の順序のカラムをすべて-1する
		return tx.Model(&domain.Column{}).
			Where("board_id = ? AND \"order\" > ?", column.BoardID, column.Order).
			Update("order", gorm.Expr("\"order\" - 1")).Error
	})
}

// UpdateOrder カラムの順序を更新します
func (r *columnRepository) UpdateOrder(id uint, newOrder int) error {
	var column domain.Column
	if err := r.db.First(&column, id).Error; err != nil {
		return err
	}

	oldOrder := column.Order
	boardID := column.BoardID

	return r.db.Transaction(func(tx *gorm.DB) error {
		if newOrder > oldOrder {
			// 右に移動する場合：間のカラムを左にシフト
			if err := tx.Model(&domain.Column{}).
				Where("board_id = ? AND \"order\" > ? AND \"order\" <= ?", boardID, oldOrder, newOrder).
				Update("order", gorm.Expr("\"order\" - 1")).Error; err != nil {
				return err
			}
		} else if newOrder < oldOrder {
			// 左に移動する場合：間のカラムを右にシフト
			if err := tx.Model(&domain.Column{}).
				Where("board_id = ? AND \"order\" >= ? AND \"order\" < ?", boardID, newOrder, oldOrder).
				Update("order", gorm.Expr("\"order\" + 1")).Error; err != nil {
				return err
			}
		}

		// 対象カラムの順序を更新
		return tx.Model(&column).Update("order", newOrder).Error
	})
}

// ReorderColumns ボード内のカラムの順序を一括更新します
func (r *columnRepository) ReorderColumns(boardID uint, columnIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, columnID := range columnIDs {
			if err := tx.Model(&domain.Column{}).
				Where("id = ? AND board_id = ?", columnID, boardID).
				Update("order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

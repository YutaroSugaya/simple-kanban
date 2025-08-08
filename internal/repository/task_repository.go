package repository

import (
	"simple-kanban/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskRepository タスクのデータアクセスを管理するインターフェース
type TaskRepository interface {
	Create(task *domain.Task) error
	GetByID(id uint) (*domain.Task, error)
	GetByColumnID(columnID uint) ([]domain.Task, error)
	GetTasksByUserID(userID uuid.UUID) ([]domain.Task, error)
	Update(task *domain.Task) error
	Delete(id uint) error
	UpdateOrder(id uint, newOrder int) error
	MoveToColumn(taskID uint, newColumnID uint, newOrder int) error
	ReorderTasksInColumn(columnID uint, taskIDs []uint) error
}

// taskRepository TaskRepositoryの実装
type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository TaskRepositoryの新しいインスタンスを作成
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

// Create 新しいタスクを作成します
func (r *taskRepository) Create(task *domain.Task) error {
	// 新しいタスクを作成する際は、そのカラムの最後の順序番号を取得して+1する
	if task.Order == 0 {
		var maxOrder int
		r.db.Model(&domain.Task{}).Where("column_id = ?", task.ColumnID).Select("COALESCE(MAX(\"order\"), 0)").Scan(&maxOrder)
		task.Order = maxOrder + 1
	}

	result := r.db.Create(task)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByID IDでタスクを取得します
func (r *taskRepository) GetByID(id uint) (*domain.Task, error) {
	var task domain.Task
	result := r.db.Preload("Column").Preload("Assignee").Where("id = ?", id).First(&task)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // タスクが見つからない場合はnilを返す
		}
		return nil, result.Error
	}
	return &task, nil
}

// GetByColumnID カラムIDでタスク一覧を取得します（順序順）
func (r *taskRepository) GetByColumnID(columnID uint) ([]domain.Task, error) {
	var tasks []domain.Task
	result := r.db.Preload("Assignee").Where("column_id = ?", columnID).Order("\"order\" ASC").Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}
	return tasks, nil
}

// GetTasksByUserID ユーザーIDでタスク一覧を取得します
func (r *taskRepository) GetTasksByUserID(userID uuid.UUID) ([]domain.Task, error) {
	var tasks []domain.Task
	result := r.db.Preload("Column").Preload("Assignee").Where("assignee_id = ?", userID).Order("\"order\" ASC").Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}
	return tasks, nil
}

// Update タスク情報を更新します
func (r *taskRepository) Update(task *domain.Task) error {
	result := r.db.Save(task)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete タスクを削除します（ソフトデリート）
func (r *taskRepository) Delete(id uint) error {
	// タスクを削除する前に、同じカラム内の他のタスクの順序を調整
	var task domain.Task
	if err := r.db.First(&task, id).Error; err != nil {
		return err
	}

	// トランザクション内で削除と順序調整を実行
	return r.db.Transaction(func(tx *gorm.DB) error {
		// タスクを削除
		if err := tx.Delete(&domain.Task{}, id).Error; err != nil {
			return err
		}

		// 削除されたタスクより後の順序のタスクをすべて-1する
		return tx.Model(&domain.Task{}).
			Where("column_id = ? AND \"order\" > ?", task.ColumnID, task.Order).
			Update("order", gorm.Expr("\"order\" - 1")).Error
	})
}

// UpdateOrder タスクの順序を更新します
func (r *taskRepository) UpdateOrder(id uint, newOrder int) error {
	var task domain.Task
	if err := r.db.First(&task, id).Error; err != nil {
		return err
	}

	oldOrder := task.Order
	columnID := task.ColumnID

	return r.db.Transaction(func(tx *gorm.DB) error {
		if newOrder > oldOrder {
			// 下に移動する場合：間のタスクを上にシフト
			if err := tx.Model(&domain.Task{}).
				Where("column_id = ? AND \"order\" > ? AND \"order\" <= ?", columnID, oldOrder, newOrder).
				Update("order", gorm.Expr("\"order\" - 1")).Error; err != nil {
				return err
			}
		} else if newOrder < oldOrder {
			// 上に移動する場合：間のタスクを下にシフト
			if err := tx.Model(&domain.Task{}).
				Where("column_id = ? AND \"order\" >= ? AND \"order\" < ?", columnID, newOrder, oldOrder).
				Update("order", gorm.Expr("\"order\" + 1")).Error; err != nil {
				return err
			}
		}

		// 対象タスクの順序を更新
		return tx.Model(&task).Update("order", newOrder).Error
	})
}

// MoveToColumn タスクを別のカラムに移動します
func (r *taskRepository) MoveToColumn(taskID uint, newColumnID uint, newOrder int) error {
	var task domain.Task
	if err := r.db.First(&task, taskID).Error; err != nil {
		return err
	}

	oldColumnID := task.ColumnID
	oldOrder := task.Order

	return r.db.Transaction(func(tx *gorm.DB) error {
		// 元のカラムで、移動したタスクより後のタスクを前にシフト
		if err := tx.Model(&domain.Task{}).
			Where("column_id = ? AND \"order\" > ?", oldColumnID, oldOrder).
			Update("order", gorm.Expr("\"order\" - 1")).Error; err != nil {
			return err
		}

		// 新しいカラムで、挿入位置以降のタスクを後ろにシフト
		if err := tx.Model(&domain.Task{}).
			Where("column_id = ? AND \"order\" >= ?", newColumnID, newOrder).
			Update("order", gorm.Expr("\"order\" + 1")).Error; err != nil {
			return err
		}

		// タスクを新しいカラムに移動
		return tx.Model(&task).Updates(map[string]interface{}{
			"column_id": newColumnID,
			"order":     newOrder,
		}).Error
	})
}

// ReorderTasksInColumn カラム内のタスクの順序を一括更新します
func (r *taskRepository) ReorderTasksInColumn(columnID uint, taskIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, taskID := range taskIDs {
			if err := tx.Model(&domain.Task{}).
				Where("id = ? AND column_id = ?", taskID, columnID).
				Update("order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

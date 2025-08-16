package service

import (
	"errors"
	"fmt"
	"time"

	"simple-kanban/internal/domain"
	"simple-kanban/internal/repository"

	"github.com/google/uuid"
)

// TaskService タスク関連のビジネスロジックを管理するインターフェース
type TaskService interface {
	CreateTask(columnID uint, userID uuid.UUID, title, description string, order int, assigneeID *uuid.UUID, dueDate *time.Time) (*domain.Task, error)
	GetTask(taskID uint, userID uuid.UUID) (*domain.Task, error)
	UpdateTask(taskID uint, userID uuid.UUID, updates map[string]interface{}) (*domain.Task, error)
	DeleteTask(taskID uint, userID uuid.UUID) error
	MoveTask(taskID uint, newColumnID uint, newOrder int, userID uuid.UUID) error
	ReorderTasks(columnID uint, taskIDs []uint, userID uuid.UUID) error
}

// taskService TaskServiceの実装
type taskService struct {
	taskRepo   repository.TaskRepository
	boardRepo  repository.BoardRepository
	columnRepo repository.ColumnRepository
}

// NewTaskService TaskServiceの新しいインスタンスを作成
func NewTaskService(taskRepo repository.TaskRepository, boardRepo repository.BoardRepository, columnRepo repository.ColumnRepository) TaskService {
	return &taskService{
		taskRepo:   taskRepo,
		boardRepo:  boardRepo,
		columnRepo: columnRepo,
	}
}

// CreateTask 新しいタスクを作成します
func (s *taskService) CreateTask(columnID uint, userID uuid.UUID, title, description string, order int, assigneeID *uuid.UUID, dueDate *time.Time) (*domain.Task, error) {
	// カラムの存在確認とボードの所有権チェック
	if err := s.checkColumnAccess(columnID, userID); err != nil {
		return nil, err
	}

	// 新しいタスクを作成
	task := &domain.Task{
		ColumnID:    columnID,
		Title:       title,
		Description: description,
		Order:       order,
		AssigneeID:  assigneeID,
		DueDate:     dueDate,
	}

	// データベースに保存
	if err := s.taskRepo.Create(task); err != nil {
		return nil, fmt.Errorf("タスク作成エラー: %w", err)
	}

	// 作成されたタスクを関連データと共に取得
	createdTask, err := s.taskRepo.GetByID(task.ID)
	if err != nil {
		return nil, fmt.Errorf("作成されたタスク取得エラー: %w", err)
	}

	return createdTask, nil
}

// GetTask タスクを取得します
func (s *taskService) GetTask(taskID uint, userID uuid.UUID) (*domain.Task, error) {
	// タスクを取得
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("タスク取得エラー: %w", err)
	}
	if task == nil {
		return nil, errors.New("タスクが見つかりません")
	}

	// ボードの所有権チェック
	if err := s.checkColumnAccess(task.ColumnID, userID); err != nil {
		return nil, err
	}

	return task, nil
}

// UpdateTask タスク情報を更新します
func (s *taskService) UpdateTask(taskID uint, userID uuid.UUID, updates map[string]interface{}) (*domain.Task, error) {
	// タスクを取得
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("タスク取得エラー: %w", err)
	}
	if task == nil {
		return nil, errors.New("タスクが見つかりません")
	}

	// ボードの所有権チェック
	if err := s.checkColumnAccess(task.ColumnID, userID); err != nil {
		return nil, err
	}

	// 更新可能なフィールドのみ処理
	if title, ok := updates["title"].(string); ok && title != "" {
		task.Title = title
	}
	if description, ok := updates["description"].(string); ok {
		task.Description = description
	}
	if assigneeID, ok := updates["assignee_id"]; ok {
		if assigneeID == nil {
			task.AssigneeID = nil
		} else if id, ok := assigneeID.(string); ok {
			if parsedID, err := uuid.Parse(id); err == nil {
				task.AssigneeID = &parsedID
			}
		}
	}
	if dueDate, ok := updates["due_date"]; ok {
		if dueDate == nil {
			task.DueDate = nil
		} else if dateStr, ok := dueDate.(string); ok {
			if parsedDate, err := time.Parse(time.RFC3339, dateStr); err == nil {
				task.DueDate = &parsedDate
			}
		} else if dateTime, ok := dueDate.(time.Time); ok {
			d := dateTime
			task.DueDate = &d
		}
	}

	// 拡張フィールドの更新処理
	if est, ok := updates["estimated_time"]; ok {
		switch v := est.(type) {
		case int:
			val := v
			task.EstimatedTime = &val
		case int32:
			val := int(v)
			task.EstimatedTime = &val
		case int64:
			val := int(v)
			task.EstimatedTime = &val
		case float64:
			val := int(v)
			task.EstimatedTime = &val
		}
	}
	if comp, ok := updates["is_completed"]; ok {
		if v, ok := comp.(bool); ok {
			task.IsCompleted = v
		}
	}
	if startVal, ok := updates["scheduled_start"]; ok {
		if startVal == nil {
			task.ScheduledStart = nil
		} else if sStr, ok := startVal.(string); ok {
			if parsed, err := time.Parse(time.RFC3339, sStr); err == nil {
				task.ScheduledStart = &parsed
			}
		} else if t, ok := startVal.(time.Time); ok {
			tmp := t
			task.ScheduledStart = &tmp
		}
	}
	if endVal, ok := updates["scheduled_end"]; ok {
		if endVal == nil {
			task.ScheduledEnd = nil
		} else if eStr, ok := endVal.(string); ok {
			if parsed, err := time.Parse(time.RFC3339, eStr); err == nil {
				task.ScheduledEnd = &parsed
			}
		} else if t, ok := endVal.(time.Time); ok {
			tmp := t
			task.ScheduledEnd = &tmp
		}
	}
	if calDate, ok := updates["calendar_date"]; ok {
		if calDate == nil {
			task.CalendarDate = nil
		} else if dStr, ok := calDate.(string); ok {
			if parsed, err := time.Parse(time.RFC3339, dStr); err == nil {
				task.CalendarDate = &parsed
			}
		} else if t, ok := calDate.(time.Time); ok {
			tmp := t
			task.CalendarDate = &tmp
		}
	}
	// データベースに保存
	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("タスク更新エラー: %w", err)
	}

	return task, nil
}

// DeleteTask タスクを削除します
func (s *taskService) DeleteTask(taskID uint, userID uuid.UUID) error {
	// タスクを取得
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return fmt.Errorf("タスク取得エラー: %w", err)
	}
	if task == nil {
		return errors.New("タスクが見つかりません")
	}

	// ボードの所有権チェック
	if err := s.checkColumnAccess(task.ColumnID, userID); err != nil {
		return err
	}

	// タスクを削除
	if err := s.taskRepo.Delete(taskID); err != nil {
		return fmt.Errorf("タスク削除エラー: %w", err)
	}

	return nil
}

// MoveTask タスクを別のカラムに移動します
func (s *taskService) MoveTask(taskID uint, newColumnID uint, newOrder int, userID uuid.UUID) error {
	// タスクを取得
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return fmt.Errorf("タスク取得エラー: %w", err)
	}
	if task == nil {
		return errors.New("タスクが見つかりません")
	}

	// 元のカラムと新しいカラムの両方のアクセス権をチェック
	if err := s.checkColumnAccess(task.ColumnID, userID); err != nil {
		return err
	}
	if err := s.checkColumnAccess(newColumnID, userID); err != nil {
		return err
	}

	// タスクを移動
	if err := s.taskRepo.MoveToColumn(taskID, newColumnID, newOrder); err != nil {
		return fmt.Errorf("タスク移動エラー: %w", err)
	}

	return nil
}

// ReorderTasks カラム内のタスクの順序を変更します
func (s *taskService) ReorderTasks(columnID uint, taskIDs []uint, userID uuid.UUID) error {
	// カラムのアクセス権をチェック
	if err := s.checkColumnAccess(columnID, userID); err != nil {
		return err
	}

	// タスクの順序を更新
	if err := s.taskRepo.ReorderTasksInColumn(columnID, taskIDs); err != nil {
		return fmt.Errorf("タスク順序更新エラー: %w", err)
	}

	return nil
}

// checkColumnAccess カラムへのアクセス権限をチェックします
func (s *taskService) checkColumnAccess(columnID uint, userID uuid.UUID) error {
	// カラムを取得
	column, err := s.columnRepo.GetByID(columnID)
	if err != nil {
		return fmt.Errorf("カラム取得エラー: %w", err)
	}
	if column == nil {
		return errors.New("カラムが見つかりません")
	}

	// ボードを取得
	board, err := s.boardRepo.GetByID(column.BoardID)
	if err != nil {
		return fmt.Errorf("ボード取得エラー: %w", err)
	}
	if board == nil {
		return errors.New("ボードが見つかりません")
	}

	// ボードの所有権をチェック
	if board.OwnerID != userID {
		return errors.New("このタスクにアクセスする権限がありません")
	}

	return nil
}

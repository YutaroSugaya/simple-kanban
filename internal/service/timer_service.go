package service

import (
	"errors"
	"simple-kanban/internal/domain"
	"simple-kanban/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TimerService タイマー関連のサービスインターフェース
type TimerService interface {
	StartTimer(userID uuid.UUID, taskID uint, duration int) (*domain.TimerSession, error)
	StopTimer(userID uuid.UUID, sessionID uint) (*domain.TimerSession, error)
	GetActiveTimer(userID uuid.UUID) (*domain.TimerSession, error)
	GetTimerHistory(userID uuid.UUID) ([]*domain.TimerSession, error)
	GetTimersByTask(taskID uint) ([]*domain.TimerSession, error)
	UpdateTaskActualTime(taskID uint) error
}

// timerService タイマーサービスの実装
type timerService struct {
	timerSessionRepo repository.TimerSessionRepository
	taskRepo         repository.TaskRepository
}

// NewTimerService タイマーサービスのコンストラクタ
func NewTimerService(
	timerSessionRepo repository.TimerSessionRepository,
	taskRepo repository.TaskRepository,
) TimerService {
	return &timerService{
		timerSessionRepo: timerSessionRepo,
		taskRepo:         taskRepo,
	}
}

// StartTimer タイマーを開始します
func (s *timerService) StartTimer(userID uuid.UUID, taskID uint, duration int) (*domain.TimerSession, error) {
	// アクティブなタイマーがないかチェック
	activeTimer, err := s.timerSessionRepo.GetActiveByUserID(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if activeTimer != nil {
		return nil, errors.New("既にアクティブなタイマーが存在します。先に停止してください")
	}

	// タスクの存在確認
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return nil, errors.New("指定されたタスクが見つかりません")
	}

	// 新しいタイマーセッションを作成
	session := &domain.TimerSession{
		TaskID:    taskID,
		UserID:    userID,
		StartTime: time.Now(),
		Duration:  duration,
		IsActive:  true,
	}

	if err := s.timerSessionRepo.Create(session); err != nil {
		return nil, err
	}

	// タスクの情報も取得して返す
	session.Task = *task

	return session, nil
}

// StopTimer タイマーを停止します
func (s *timerService) StopTimer(userID uuid.UUID, sessionID uint) (*domain.TimerSession, error) {
	// セッションを取得
	session, err := s.timerSessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, err
	}

	// ユーザー権限チェック
	if session.UserID != userID {
		return nil, errors.New("このタイマーを操作する権限がありません")
	}

	// アクティブ状態チェック
	if !session.IsActive {
		return nil, errors.New("このタイマーは既に停止されています")
	}

	// タイマーを停止
	now := time.Now()
	actualDuration := int(now.Sub(session.StartTime).Seconds())

	session.EndTime = &now
	session.Duration = actualDuration
	session.IsActive = false

	if err := s.timerSessionRepo.Update(session); err != nil {
		return nil, err
	}

	// タスクの実際の時間を更新
	if err := s.UpdateTaskActualTime(session.TaskID); err != nil {
		// ログに記録するが、セッション更新は成功とする
		// TODO: ログ機能実装時にログ出力を追加
	}

	return session, nil
}

// GetActiveTimer アクティブなタイマーを取得します
func (s *timerService) GetActiveTimer(userID uuid.UUID) (*domain.TimerSession, error) {
	return s.timerSessionRepo.GetActiveByUserID(userID)
}

// GetTimerHistory タイマー履歴を取得します
func (s *timerService) GetTimerHistory(userID uuid.UUID) ([]*domain.TimerSession, error) {
	return s.timerSessionRepo.GetByUserID(userID)
}

// GetTimersByTask タスクのタイマー履歴を取得します
func (s *timerService) GetTimersByTask(taskID uint) ([]*domain.TimerSession, error) {
	return s.timerSessionRepo.GetByTaskID(taskID)
}

// UpdateTaskActualTime タスクの実際の時間を更新します
func (s *timerService) UpdateTaskActualTime(taskID uint) error {
	// タスクを取得
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return err
	}

	// タスクの全タイマーセッションを取得
	sessions, err := s.timerSessionRepo.GetByTaskID(taskID)
	if err != nil {
		return err
	}

	// 合計時間を計算（分単位）
	totalSeconds := 0
	for _, session := range sessions {
		if !session.IsActive {
			totalSeconds += session.Duration
		}
	}

	totalMinutes := totalSeconds / 60
	task.ActualTime = &totalMinutes

	return s.taskRepo.Update(task)
}

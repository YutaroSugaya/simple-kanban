package repository

import (
	"simple-kanban/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TimerSessionRepository タイマーセッションのリポジトリインターフェース
type TimerSessionRepository interface {
	Create(session *domain.TimerSession) error
	GetByID(id uint) (*domain.TimerSession, error)
	GetByTaskID(taskID uint) ([]*domain.TimerSession, error)
	GetByUserID(userID uuid.UUID) ([]*domain.TimerSession, error)
	GetActiveByUserID(userID uuid.UUID) (*domain.TimerSession, error)
	Update(session *domain.TimerSession) error
	Delete(id uint) error
}

// timerSessionRepository タイマーセッションリポジトリの実装
type timerSessionRepository struct {
	db *gorm.DB
}

// NewTimerSessionRepository タイマーセッションリポジトリのコンストラクタ
func NewTimerSessionRepository(db *gorm.DB) TimerSessionRepository {
	return &timerSessionRepository{db: db}
}

// Create タイマーセッションを作成します
func (r *timerSessionRepository) Create(session *domain.TimerSession) error {
	return r.db.Create(session).Error
}

// GetByID IDでタイマーセッションを取得します
func (r *timerSessionRepository) GetByID(id uint) (*domain.TimerSession, error) {
	var session domain.TimerSession
	err := r.db.Preload("Task").Preload("User").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByTaskID タスクIDでタイマーセッションを取得します
func (r *timerSessionRepository) GetByTaskID(taskID uint) ([]*domain.TimerSession, error) {
	var sessions []*domain.TimerSession
	err := r.db.Where("task_id = ?", taskID).
		Preload("Task").Preload("User").
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// GetByUserID ユーザーIDでタイマーセッションを取得します
func (r *timerSessionRepository) GetByUserID(userID uuid.UUID) ([]*domain.TimerSession, error) {
	var sessions []*domain.TimerSession
	err := r.db.Where("user_id = ?", userID).
		Preload("Task").Preload("User").
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// GetActiveByUserID ユーザーのアクティブなタイマーセッションを取得します
func (r *timerSessionRepository) GetActiveByUserID(userID uuid.UUID) (*domain.TimerSession, error) {
	var session domain.TimerSession
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Preload("Task").Preload("User").
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update タイマーセッションを更新します
func (r *timerSessionRepository) Update(session *domain.TimerSession) error {
	return r.db.Save(session).Error
}

// Delete タイマーセッションを削除します
func (r *timerSessionRepository) Delete(id uint) error {
	return r.db.Delete(&domain.TimerSession{}, id).Error
}

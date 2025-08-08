package service

import (
	"errors"
	"log"
	"simple-kanban/internal/domain"
	"simple-kanban/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CalendarService カレンダー関連のサービスインターフェース
type CalendarService interface {
	// カレンダー設定関連
	CreateCalendarSettings(userID uuid.UUID, settings *domain.CalendarSettings) error
	GetCalendarSettings(userID uuid.UUID) (*domain.CalendarSettings, error)
	UpdateCalendarSettings(userID uuid.UUID, settings *domain.CalendarSettings) error

	// カレンダーイベント関連
	CreateEvent(userID uuid.UUID, event *domain.CalendarEvent) error
	GetEventsByDateRange(userID uuid.UUID, start, end time.Time) ([]*domain.CalendarEvent, error)
	UpdateEvent(userID uuid.UUID, eventID uint, event *domain.CalendarEvent) error
	DeleteEvent(userID uuid.UUID, eventID uint) error

	// タスクからカレンダーイベント作成
	CreateEventFromTask(userID uuid.UUID, task *domain.Task, start, end time.Time) error
	UpdateTaskSchedule(userID uuid.UUID, taskID uint, start, end time.Time) error
}

// calendarService カレンダーサービスの実装
type calendarService struct {
	calendarSettingsRepo repository.CalendarSettingsRepository
	calendarEventRepo    repository.CalendarEventRepository
	taskRepo             repository.TaskRepository
}

// NewCalendarService カレンダーサービスのコンストラクタ
func NewCalendarService(
	calendarSettingsRepo repository.CalendarSettingsRepository,
	calendarEventRepo repository.CalendarEventRepository,
	taskRepo repository.TaskRepository,
) CalendarService {
	return &calendarService{
		calendarSettingsRepo: calendarSettingsRepo,
		calendarEventRepo:    calendarEventRepo,
		taskRepo:             taskRepo,
	}
}

// CreateCalendarSettings カレンダー設定を作成します
func (s *calendarService) CreateCalendarSettings(userID uuid.UUID, settings *domain.CalendarSettings) error {
	// 既存の設定があるかチェック
	existing, err := s.calendarSettingsRepo.GetByUserID(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if existing != nil {
		return errors.New("カレンダー設定は既に存在します")
	}

	settings.UserID = userID
	return s.calendarSettingsRepo.Create(settings)
}

// GetCalendarSettings カレンダー設定を取得します
func (s *calendarService) GetCalendarSettings(userID uuid.UUID) (*domain.CalendarSettings, error) {
	settings, err := s.calendarSettingsRepo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// デフォルト設定を作成
			defaultSettings := &domain.CalendarSettings{
				UserID:           userID,
				WeekdayStartTime: "09:00",
				WeekdayEndTime:   "18:00",
				WeekendStartTime: "10:00",
				WeekendEndTime:   "16:00",
				TimeSlotDuration: 10,
			}
			if createErr := s.calendarSettingsRepo.Create(defaultSettings); createErr != nil {
				return nil, createErr
			}
			return defaultSettings, nil
		}
		return nil, err
	}
	return settings, nil
}

// UpdateCalendarSettings カレンダー設定を更新します
func (s *calendarService) UpdateCalendarSettings(userID uuid.UUID, settings *domain.CalendarSettings) error {
	existing, err := s.calendarSettingsRepo.GetByUserID(userID)
	if err != nil {
		return err
	}

	// 更新フィールドを設定
	existing.WeekdayStartTime = settings.WeekdayStartTime
	existing.WeekdayEndTime = settings.WeekdayEndTime
	existing.WeekendStartTime = settings.WeekendStartTime
	existing.WeekendEndTime = settings.WeekendEndTime
	existing.TimeSlotDuration = settings.TimeSlotDuration

	return s.calendarSettingsRepo.Update(existing)
}

// CreateEvent カレンダーイベントを作成します
func (s *calendarService) CreateEvent(userID uuid.UUID, event *domain.CalendarEvent) error {
	event.UserID = userID
	return s.calendarEventRepo.Create(event)
}

// GetEventsByDateRange 指定期間のカレンダーイベントを取得します
func (s *calendarService) GetEventsByDateRange(userID uuid.UUID, start, end time.Time) ([]*domain.CalendarEvent, error) {
	return s.calendarEventRepo.GetByUserIDAndDateRange(userID, start, end)
}

// UpdateEvent カレンダーイベントを更新します
func (s *calendarService) UpdateEvent(userID uuid.UUID, eventID uint, event *domain.CalendarEvent) error {
	existing, err := s.calendarEventRepo.GetByID(eventID)
	if err != nil {
		return err
	}

	// ユーザー権限チェック
	if existing.UserID != userID {
		return errors.New("このイベントを更新する権限がありません")
	}

	// 更新フィールドを設定
	existing.Title = event.Title
	existing.Start = event.Start
	existing.End = event.End
	existing.Color = event.Color

	return s.calendarEventRepo.Update(existing)
}

// DeleteEvent カレンダーイベントを削除します
func (s *calendarService) DeleteEvent(userID uuid.UUID, eventID uint) error {
	existing, err := s.calendarEventRepo.GetByID(eventID)
	if err != nil {
		return err
	}

	// ユーザー権限チェック
	if existing.UserID != userID {
		return errors.New("このイベントを削除する権限がありません")
	}

	return s.calendarEventRepo.Delete(eventID)
}

// CreateEventFromTask タスクからカレンダーイベントを作成します
func (s *calendarService) CreateEventFromTask(userID uuid.UUID, task *domain.Task, start, end time.Time) error {
	log.Printf("CreateEventFromTask: 処理開始 - UserID: %s, TaskID: %d, Start: %v, End: %v",
		userID, task.ID, start, end)

	log.Printf("CreateEventFromTask: 新規イベント作成 - TaskID: %d", task.ID)

	event := &domain.CalendarEvent{
		UserID:      userID,
		TaskID:      &task.ID,
		Title:       task.Title,
		Start:       start,
		End:         end,
		Color:       "#10B981", // タスクベースイベントの色
		IsTaskBased: true,
	}

	log.Printf("CreateEventFromTask: カレンダーイベント作成開始 - TaskID: %d", task.ID)
	if err := s.calendarEventRepo.Create(event); err != nil {
		log.Printf("CreateEventFromTask: カレンダーイベント作成エラー - TaskID: %d, エラー: %v", task.ID, err)
		return err
	}

	log.Printf("CreateEventFromTask: 処理完了 - UserID: %s, TaskID: %d", userID, task.ID)
	return nil
}

// UpdateTaskSchedule タスクのスケジュールを更新します
func (s *calendarService) UpdateTaskSchedule(userID uuid.UUID, taskID uint, start, end time.Time) error {
	// タスクを取得
	task, err := s.taskRepo.GetByID(taskID)
	if err != nil {
		return err
	}

	// カレンダーイベントを取得
	event, err := s.calendarEventRepo.GetByTaskID(taskID)
	if err != nil {
		return errors.New("タスクに関連するカレンダーイベントが見つかりません")
	}

	// ユーザー権限チェック
	if event.UserID != userID {
		return errors.New("このタスクを更新する権限がありません")
	}

	// タスクのスケジュール更新
	task.ScheduledStart = &start
	task.ScheduledEnd = &end
	task.CalendarDate = &start

	// イベントの時間更新
	event.Start = start
	event.End = end

	// 両方を更新
	if err := s.taskRepo.Update(task); err != nil {
		return err
	}

	return s.calendarEventRepo.Update(event)
}

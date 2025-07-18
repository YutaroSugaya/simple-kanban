package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"simple-kanban/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskService モックサービス
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(columnID uint, userID uuid.UUID, title, description string, order int, assigneeID *uuid.UUID, dueDate *time.Time) (*domain.Task, error) {
	args := m.Called(columnID, userID, title, description, order, assigneeID, dueDate)
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) GetTask(taskID uint, userID uuid.UUID) (*domain.Task, error) {
	args := m.Called(taskID, userID)
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) UpdateTask(taskID uint, userID uuid.UUID, updates map[string]interface{}) (*domain.Task, error) {
	args := m.Called(taskID, userID, updates)
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) DeleteTask(taskID uint, userID uuid.UUID) error {
	args := m.Called(taskID, userID)
	return args.Error(0)
}

func (m *MockTaskService) MoveTask(taskID uint, newColumnID uint, newOrder int, userID uuid.UUID) error {
	args := m.Called(taskID, newColumnID, newOrder, userID)
	return args.Error(0)
}

func (m *MockTaskService) ReorderTasks(columnID uint, taskIDs []uint, userID uuid.UUID) error {
	args := m.Called(columnID, taskIDs, userID)
	return args.Error(0)
}

// テスト用のヘルパー関数
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func createTestRequest(method, url string, body interface{}) (*http.Request, error) {
	var req *http.Request
	var err error

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// UpdateTaskのテスト
func TestUpdateTask_Success(t *testing.T) {
	// テストデータ準備
	mockService := new(MockTaskService)
	handler := NewTaskHandler(mockService)

	userID := uuid.New()
	taskID := uint(1)

	// 期待されるリクエスト
	requestBody := UpdateTaskRequest{
		Title:       stringPtr("更新されたタスク"),
		Description: stringPtr("更新された説明"),
		DueDate:     stringPtr("2024-01-15"),
	}

	// 期待されるレスポンス
	expectedTask := &domain.Task{
		ID:          taskID,
		Title:       "更新されたタスク",
		Description: "更新された説明",
		DueDate:     parseTime("2024-01-15"),
	}

	// モックの設定
	mockService.On("UpdateTask", taskID, userID, mock.AnythingOfType("map[string]interface {}")).Return(expectedTask, nil)

	// テスト実行
	router := setupTestRouter()
	router.PUT("/tasks/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.UpdateTask(c)
	})

	req, err := createTestRequest("PUT", "/tasks/1", requestBody)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// アサーション
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "更新されたタスク", response["title"])
	assert.Equal(t, "更新された説明", response["description"])

	mockService.AssertExpectations(t)
}

// UpdateTaskのバリデーションエラーテスト
func TestUpdateTask_ValidationError(t *testing.T) {
	mockService := new(MockTaskService)
	handler := NewTaskHandler(mockService)

	userID := uuid.New()

	// 無効なリクエスト（空のタイトル）
	requestBody := UpdateTaskRequest{
		Title: stringPtr(""), // 空文字列（バリデーションエラー）
	}

	router := setupTestRouter()
	router.PUT("/tasks/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.UpdateTask(c)
	})

	req, err := createTestRequest("PUT", "/tasks/1", requestBody)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// アサーション
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["error"], "バリデーションエラー")
}

// UpdateTaskのJSONバインドエラーテスト
func TestUpdateTask_JSONBindError(t *testing.T) {
	mockService := new(MockTaskService)
	handler := NewTaskHandler(mockService)

	userID := uuid.New()

	router := setupTestRouter()
	router.PUT("/tasks/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.UpdateTask(c)
	})

	// 不正なJSONを送信
	req, err := http.NewRequest("PUT", "/tasks/1", bytes.NewBufferString(`{"title": "test"`))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// アサーション
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["error"], "不正なリクエスト形式")
}

// ベンチマークテスト
func BenchmarkUpdateTask(b *testing.B) {
	mockService := new(MockTaskService)
	handler := NewTaskHandler(mockService)

	userID := uuid.New()
	taskID := uint(1)

	requestBody := UpdateTaskRequest{
		Title:       stringPtr("ベンチマークテスト"),
		Description: stringPtr("ベンチマーク用の説明"),
		DueDate:     stringPtr("2024-01-15"),
	}

	expectedTask := &domain.Task{
		ID:          taskID,
		Title:       "ベンチマークテスト",
		Description: "ベンチマーク用の説明",
		DueDate:     parseTime("2024-01-15"),
	}

	mockService.On("UpdateTask", taskID, userID, mock.AnythingOfType("map[string]interface {}")).Return(expectedTask, nil)

	router := setupTestRouter()
	router.PUT("/tasks/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.UpdateTask(c)
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := createTestRequest("PUT", "/tasks/1", requestBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// ヘルパー関数
func stringPtr(s string) *string {
	return &s
}

func parseTime(dateStr string) *time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return &t
}

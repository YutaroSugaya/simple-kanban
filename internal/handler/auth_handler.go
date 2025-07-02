package handler

import (
	"net/http"

	"simple-kanban/internal/service"
	"simple-kanban/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// AuthHandler 認証関連のHTTPハンドラ
type AuthHandler struct {
	userService service.UserService
	validator   *validator.Validate
}

// NewAuthHandler AuthHandlerの新しいインスタンスを作成
func NewAuthHandler(userService service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		validator:   validator.New(),
	}
}

// RegisterRequest ユーザー登録リクエスト構造体
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest ログインリクエスト構造体
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse 認証レスポンス構造体
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// UserResponse ユーザー情報レスポンス構造体
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Register ユーザー登録ハンドラ
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なリクエスト形式です",
		})
		return
	}

	// バリデーション
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "バリデーションエラー",
			"details": err.Error(),
		})
		return
	}

	// ユーザー登録処理
	user, token, err := h.userService.Register(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// レスポンスを返す
	response := AuthResponse{
		User: UserResponse{
			ID:    user.ID.String(),
			Email: user.Email,
		},
		Token: token,
	}

	c.JSON(http.StatusCreated, response)
}

// Login ログインハンドラ
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不正なリクエスト形式です",
		})
		return
	}

	// バリデーション
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "バリデーションエラー",
			"details": err.Error(),
		})
		return
	}

	// ログイン処理
	user, token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// レスポンスを返す
	response := AuthResponse{
		User: UserResponse{
			ID:    user.ID.String(),
			Email: user.Email,
		},
		Token: token,
	}

	c.JSON(http.StatusOK, response)
}

// Profile ユーザープロフィール取得ハンドラ
// GET /api/v1/auth/profile
func (h *AuthHandler) Profile(c *gin.Context) {
	// JWT認証ミドルウェアからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "認証情報が取得できません",
		})
		return
	}

	// ユーザー情報を取得
	user, err := h.userService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// レスポンスを返す
	response := UserResponse{
		ID:    user.ID.String(),
		Email: user.Email,
	}

	c.JSON(http.StatusOK, gin.H{
		"user": response,
	})
}

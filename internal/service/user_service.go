package service

import (
	"errors"
	"fmt"

	"simple-kanban/config"
	"simple-kanban/internal/domain"
	"simple-kanban/internal/repository"
	"simple-kanban/pkg/middleware"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserService ユーザー関連のビジネスロジックを管理するインターフェース
type UserService interface {
	Register(email, password string) (*domain.User, string, error)
	Login(email, password string) (*domain.User, string, error)
	GetProfile(userID uuid.UUID) (*domain.User, error)
	UpdateProfile(userID uuid.UUID, updates map[string]interface{}) (*domain.User, error)
}

// userService UserServiceの実装
type userService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

// NewUserService UserServiceの新しいインスタンスを作成
func NewUserService(userRepo repository.UserRepository, cfg *config.Config) UserService {
	return &userService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register 新しいユーザーを登録します
func (s *userService) Register(email, password string) (*domain.User, string, error) {
	// メールアドレスの重複チェック
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("ユーザー存在確認エラー: %w", err)
	}
	if existingUser != nil {
		return nil, "", errors.New("このメールアドレスは既に使用されています")
	}

	// パスワードをハッシュ化
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("パスワードハッシュ化エラー: %w", err)
	}

	// 新しいユーザーを作成
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hashedPassword,
	}

	// データベースに保存
	if err := s.userRepo.Create(user); err != nil {
		return nil, "", fmt.Errorf("ユーザー作成エラー: %w", err)
	}

	// JWTトークンを生成
	token, err := middleware.GenerateToken(user.ID, user.Email, s.cfg)
	if err != nil {
		return nil, "", fmt.Errorf("トークン生成エラー: %w", err)
	}

	return user, token, nil
}

// Login ユーザーのログイン認証を行います
func (s *userService) Login(email, password string) (*domain.User, string, error) {
	// メールアドレスでユーザーを検索
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("ユーザー検索エラー: %w", err)
	}
	if user == nil {
		return nil, "", errors.New("メールアドレスまたはパスワードが正しくありません")
	}

	// パスワードを検証
	if !s.checkPasswordHash(password, user.PasswordHash) {
		return nil, "", errors.New("メールアドレスまたはパスワードが正しくありません")
	}

	// JWTトークンを生成
	token, err := middleware.GenerateToken(user.ID, user.Email, s.cfg)
	if err != nil {
		return nil, "", fmt.Errorf("トークン生成エラー: %w", err)
	}

	return user, token, nil
}

// GetProfile ユーザーのプロフィール情報を取得します
func (s *userService) GetProfile(userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー取得エラー: %w", err)
	}
	if user == nil {
		return nil, errors.New("ユーザーが見つかりません")
	}
	return user, nil
}

// UpdateProfile ユーザーのプロフィール情報を更新します
func (s *userService) UpdateProfile(userID uuid.UUID, updates map[string]interface{}) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー取得エラー: %w", err)
	}
	if user == nil {
		return nil, errors.New("ユーザーが見つかりません")
	}

	// 更新可能なフィールドのみ処理
	if email, ok := updates["email"].(string); ok && email != "" {
		// メールアドレスの重複チェック
		existingUser, err := s.userRepo.GetByEmail(email)
		if err != nil {
			return nil, fmt.Errorf("メール重複確認エラー: %w", err)
		}
		if existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("このメールアドレスは既に使用されています")
		}
		user.Email = email
	}

	// パスワード更新の場合はハッシュ化
	if password, ok := updates["password"].(string); ok && password != "" {
		hashedPassword, err := s.hashPassword(password)
		if err != nil {
			return nil, fmt.Errorf("パスワードハッシュ化エラー: %w", err)
		}
		user.PasswordHash = hashedPassword
	}

	// データベースに保存
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("ユーザー更新エラー: %w", err)
	}

	return user, nil
}

// hashPassword パスワードをハッシュ化します
func (s *userService) hashPassword(password string) (string, error) {
	// bcryptでパスワードをハッシュ化（コスト10）
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// checkPasswordHash パスワードとハッシュを照合します
func (s *userService) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

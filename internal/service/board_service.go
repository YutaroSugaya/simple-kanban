package service

import (
	"errors"
	"fmt"

	"simple-kanban/internal/domain"
	"simple-kanban/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BoardService ボード関連のビジネスロジックを管理するインターフェース
type BoardService interface {
	CreateBoard(ownerID uuid.UUID, name string) (*domain.Board, error)
	GetUserBoards(userID uuid.UUID) ([]domain.Board, error)
	GetBoardWithColumns(boardID uint, userID uuid.UUID) (*domain.Board, error)
	UpdateBoard(boardID uint, userID uuid.UUID, updates map[string]interface{}) (*domain.Board, error)
	DeleteBoard(boardID uint, userID uuid.UUID) error
	CheckBoardOwnership(boardID uint, userID uuid.UUID) error
}

// boardService BoardServiceの実装
type boardService struct {
	boardRepo  repository.BoardRepository
	columnRepo repository.ColumnRepository // カラムリポジトリ（作成予定）
	db         *gorm.DB
}

// NewBoardService BoardServiceの新しいインスタンスを作成
func NewBoardService(boardRepo repository.BoardRepository, db *gorm.DB) BoardService {
	return &boardService{
		boardRepo: boardRepo,
		db:        db,
	}
}

// CreateBoard 新しいボードを作成します
func (s *boardService) CreateBoard(ownerID uuid.UUID, name string) (*domain.Board, error) {
	// 新しいボードを作成
	board := &domain.Board{
		Name:    name,
		OwnerID: ownerID,
	}

	// トランザクション内でボードとデフォルトカラムを作成
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// ボードを作成
		if err := s.boardRepo.Create(board); err != nil {
			return fmt.Errorf("ボード作成エラー: %w", err)
		}

		// デフォルトのカラムを作成
		defaultColumns := []domain.Column{
			{BoardID: board.ID, Title: "To Do", Order: 1},
			{BoardID: board.ID, Title: "In Progress", Order: 2},
			{BoardID: board.ID, Title: "Done", Order: 3},
		}

		for _, column := range defaultColumns {
			if err := tx.Create(&column).Error; err != nil {
				return fmt.Errorf("デフォルトカラム作成エラー: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return board, nil
}

// GetUserBoards ユーザーが所有するボード一覧を取得します
func (s *boardService) GetUserBoards(userID uuid.UUID) ([]domain.Board, error) {
	boards, err := s.boardRepo.GetByOwnerID(userID)
	if err != nil {
		return nil, fmt.Errorf("ボード取得エラー: %w", err)
	}
	return boards, nil
}

// GetBoardWithColumns ボードをカラムとタスク情報付きで取得します
func (s *boardService) GetBoardWithColumns(boardID uint, userID uuid.UUID) (*domain.Board, error) {
	// ボードの所有権をチェック
	if err := s.CheckBoardOwnership(boardID, userID); err != nil {
		return nil, err
	}

	// ボードをカラム情報付きで取得
	board, err := s.boardRepo.GetByIDWithColumns(boardID)
	if err != nil {
		return nil, fmt.Errorf("ボード取得エラー: %w", err)
	}
	if board == nil {
		return nil, errors.New("ボードが見つかりません")
	}

	return board, nil
}

// UpdateBoard ボード情報を更新します
func (s *boardService) UpdateBoard(boardID uint, userID uuid.UUID, updates map[string]interface{}) (*domain.Board, error) {
	// ボードの所有権をチェック
	if err := s.CheckBoardOwnership(boardID, userID); err != nil {
		return nil, err
	}

	// ボードを取得
	board, err := s.boardRepo.GetByID(boardID)
	if err != nil {
		return nil, fmt.Errorf("ボード取得エラー: %w", err)
	}
	if board == nil {
		return nil, errors.New("ボードが見つかりません")
	}

	// 更新可能なフィールドのみ処理
	if name, ok := updates["name"].(string); ok && name != "" {
		board.Name = name
	}

	// データベースに保存
	if err := s.boardRepo.Update(board); err != nil {
		return nil, fmt.Errorf("ボード更新エラー: %w", err)
	}

	return board, nil
}

// DeleteBoard ボードを削除します
func (s *boardService) DeleteBoard(boardID uint, userID uuid.UUID) error {
	// ボードの所有権をチェック
	if err := s.CheckBoardOwnership(boardID, userID); err != nil {
		return err
	}

	// ボードを削除（カスケード削除でカラムとタスクも削除される）
	if err := s.boardRepo.Delete(boardID); err != nil {
		return fmt.Errorf("ボード削除エラー: %w", err)
	}

	return nil
}

// CheckBoardOwnership ボードの所有権をチェックします
func (s *boardService) CheckBoardOwnership(boardID uint, userID uuid.UUID) error {
	board, err := s.boardRepo.GetByID(boardID)
	if err != nil {
		return fmt.Errorf("ボード取得エラー: %w", err)
	}
	if board == nil {
		return errors.New("ボードが見つかりません")
	}
	if board.OwnerID != userID {
		return errors.New("このボードにアクセスする権限がありません")
	}
	return nil
}

package usecase

import (
	"fmt"
	"time"

	"book-manager/internal/model"
	"book-manager/internal/repository"
	"github.com/go-playground/validator/v10"
)

// BookUsecase は書籍管理のビジネスロジックを定義するインターフェース
type BookUsecase interface {
	CreateBook(req *model.CreateBookRequest) (*model.Book, error)
	GetBook(id int) (*model.Book, error)
	ListBooks(filter *model.BookFilter, page, limit int) ([]*model.Book, int, error)
	UpdateBook(id int, req *model.UpdateBookRequest) (*model.Book, error)
	DeleteBook(id int) error
	StartReading(id int) (*model.Book, error)
	FinishReading(id int, rating *int) (*model.Book, error)
	GetStatistics() (*BookStatistics, error)
}

// BookStatistics は書籍の統計情報を表す構造体
type BookStatistics struct {
	TotalBooks       int `json:"total_books"`
	NotStartedBooks  int `json:"not_started_books"`
	ReadingBooks     int `json:"reading_books"`
	CompletedBooks   int `json:"completed_books"`
	DroppedBooks     int `json:"dropped_books"`
	TotalSpent       int `json:"total_spent"`
	AverageRating    *float64 `json:"average_rating"`
	BooksThisMonth   int `json:"books_this_month"`
	CompletedThisMonth int `json:"completed_this_month"`
}

// bookUsecase はBookUsecaseの実装
type bookUsecase struct {
	bookRepo  repository.BookRepository
	validator *validator.Validate
}

// NewBookUsecase は新しいBookUsecaseを作成する
func NewBookUsecase(bookRepo repository.BookRepository) BookUsecase {
	return &bookUsecase{
		bookRepo:  bookRepo,
		validator: validator.New(),
	}
}

// CreateBook は新しい書籍を作成する
func (u *bookUsecase) CreateBook(req *model.CreateBookRequest) (*model.Book, error) {
	if err := u.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("入力データが無効です: %w", err)
	}

	// 購入日が未来でないことを確認
	if req.PurchaseDate.After(time.Now()) {
		return nil, fmt.Errorf("購入日は現在以前の日付を指定してください")
	}

	return u.bookRepo.Create(req)
}

// GetBook は指定されたIDの書籍を取得する
func (u *bookUsecase) GetBook(id int) (*model.Book, error) {
	if id <= 0 {
		return nil, fmt.Errorf("無効な書籍IDです: %d", id)
	}

	return u.bookRepo.GetByID(id)
}

// ListBooks は書籍一覧を取得する（ページネーション対応）
func (u *bookUsecase) ListBooks(filter *model.BookFilter, page, limit int) ([]*model.Book, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20 // デフォルト値
	}

	offset := (page - 1) * limit

	books, err := u.bookRepo.List(filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := u.bookRepo.Count(filter)
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

// UpdateBook は書籍情報を更新する
func (u *bookUsecase) UpdateBook(id int, req *model.UpdateBookRequest) (*model.Book, error) {
	if id <= 0 {
		return nil, fmt.Errorf("無効な書籍IDです: %d", id)
	}

	// 既存の書籍が存在するか確認
	_, err := u.bookRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 評価が1-5の範囲内かチェック
	if req.Rating != nil && (*req.Rating < 1 || *req.Rating > 5) {
		return nil, fmt.Errorf("評価は1-5の範囲で入力してください: %d", *req.Rating)
	}

	return u.bookRepo.Update(id, req)
}

// DeleteBook は書籍を削除する
func (u *bookUsecase) DeleteBook(id int) error {
	if id <= 0 {
		return fmt.Errorf("無効な書籍IDです: %d", id)
	}

	// 既存の書籍が存在するか確認
	_, err := u.bookRepo.GetByID(id)
	if err != nil {
		return err
	}

	return u.bookRepo.Delete(id)
}

// StartReading は読書を開始する
func (u *bookUsecase) StartReading(id int) (*model.Book, error) {
	if id <= 0 {
		return nil, fmt.Errorf("無効な書籍IDです: %d", id)
	}

	book, err := u.bookRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 既に読書中または読了している場合はエラー
	if book.Status == model.StatusReading {
		return nil, fmt.Errorf("この書籍は既に読書中です")
	}
	if book.Status == model.StatusCompleted {
		return nil, fmt.Errorf("この書籍は既に読了済みです")
	}

	status := model.StatusReading
	now := time.Now()
	updateReq := &model.UpdateBookRequest{
		Status:        &status,
		StartReadDate: &now,
	}

	return u.bookRepo.Update(id, updateReq)
}

// FinishReading は読書を完了する
func (u *bookUsecase) FinishReading(id int, rating *int) (*model.Book, error) {
	if id <= 0 {
		return nil, fmt.Errorf("無効な書籍IDです: %d", id)
	}

	book, err := u.bookRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 読書中でない場合はエラー
	if book.Status != model.StatusReading {
		return nil, fmt.Errorf("この書籍は読書中ではありません")
	}

	// 評価のバリデーション
	if rating != nil && (*rating < 1 || *rating > 5) {
		return nil, fmt.Errorf("評価は1-5の範囲で入力してください: %d", *rating)
	}

	status := model.StatusCompleted
	now := time.Now()
	updateReq := &model.UpdateBookRequest{
		Status:      &status,
		EndReadDate: &now,
		Rating:      rating,
	}

	return u.bookRepo.Update(id, updateReq)
}

// GetStatistics は書籍の統計情報を取得する
func (u *bookUsecase) GetStatistics() (*BookStatistics, error) {
	stats := &BookStatistics{}

	// 全書籍数
	total, err := u.bookRepo.Count(nil)
	if err != nil {
		return nil, fmt.Errorf("統計情報の取得に失敗しました: %w", err)
	}
	stats.TotalBooks = total

	// ステータス別の書籍数
	statusCounts := map[model.ReadingStatus]*int{
		model.StatusNotStarted: &stats.NotStartedBooks,
		model.StatusReading:    &stats.ReadingBooks,
		model.StatusCompleted:  &stats.CompletedBooks,
		model.StatusDropped:    &stats.DroppedBooks,
	}

	for status, countPtr := range statusCounts {
		filter := &model.BookFilter{Status: &status}
		count, err := u.bookRepo.Count(filter)
		if err != nil {
			return nil, fmt.Errorf("ステータス別統計の取得に失敗しました: %w", err)
		}
		*countPtr = count
	}

	// 全書籍を取得して金額や評価を計算
	allBooks, err := u.bookRepo.List(nil, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("書籍一覧の取得に失敗しました: %w", err)
	}

	totalSpent := 0
	ratingSum := 0
	ratingCount := 0
	booksThisMonth := 0
	completedThisMonth := 0
	
	now := time.Now()
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	for _, book := range allBooks {
		// 合計支出額
		totalSpent += book.PurchasePrice

		// 評価の平均
		if book.Rating != nil {
			ratingSum += *book.Rating
			ratingCount++
		}

		// 今月購入した書籍数
		if book.PurchaseDate.After(thisMonthStart) || book.PurchaseDate.Equal(thisMonthStart) {
			booksThisMonth++
		}

		// 今月完了した書籍数
		if book.EndReadDate != nil && 
		   (book.EndReadDate.After(thisMonthStart) || book.EndReadDate.Equal(thisMonthStart)) &&
		   book.Status == model.StatusCompleted {
			completedThisMonth++
		}
	}

	stats.TotalSpent = totalSpent
	stats.BooksThisMonth = booksThisMonth
	stats.CompletedThisMonth = completedThisMonth

	if ratingCount > 0 {
		avg := float64(ratingSum) / float64(ratingCount)
		stats.AverageRating = &avg
	}

	return stats, nil
}
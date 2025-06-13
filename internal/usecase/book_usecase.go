// usecaseパッケージ：ビジネスロジック（業務処理のルール）を担当するファイル
// UseCase層は「ビジネスロジック層」とも呼ばれ、アプリ固有の処理ルールを定義
package usecase

// import：他のパッケージ（機能）を使うための宣言
import (
	"fmt"                                       // 文字列フォーマット（エラーメッセージ作成など）
	"time"                                      // 時間関連の処理

	"book-manager/internal/model"                // 自作のデータ構造定義
	"book-manager/internal/repository"           // 自作のデータアクセス層
	"github.com/go-playground/validator/v10"   // 入力データのバリデーション（検証）ライブラリ
)

// BookUsecase は書籍管理のビジネスロジックを定義するインターフェース
// ビジネスロジック：アプリの業務ルール（例：評価は1-5点、読書中は再開始不可など）
type BookUsecase interface {
	CreateBook(req *model.CreateBookRequest) (*model.Book, error)            // 新しい書籍を作成
	GetBook(id int) (*model.Book, error)                                     // IDで書籍を1件取得
	ListBooks(filter *model.BookFilter, page, limit int) ([]*model.Book, int, error) // 書籍一覧をページング付きで取得
	UpdateBook(id int, req *model.UpdateBookRequest) (*model.Book, error)    // 書籍情報を更新
	DeleteBook(id int) error                                                 // 書籍を削除
	StartReading(id int) (*model.Book, error)                                // 読書を開始（ステータス変更）
	FinishReading(id int, rating *int) (*model.Book, error)                  // 読書を完了（評価付き）
	GetStatistics() (*BookStatistics, error)                                // 統計情報（合計金額、平均評価など）を取得
}

// BookStatistics は書籍の統計情報を表す構造体
// 統計情報：データを集計して得られる情報（合計、平均、件数など）
type BookStatistics struct {
	TotalBooks         int      `json:"total_books"`          // 総書籍数
	NotStartedBooks    int      `json:"not_started_books"`    // 未読の書籍数
	ReadingBooks       int      `json:"reading_books"`        // 読書中の書籍数
	CompletedBooks     int      `json:"completed_books"`      // 読了済みの書籍数
	DroppedBooks       int      `json:"dropped_books"`        // 中断した書籍数
	TotalSpent         int      `json:"total_spent"`          // 総支出金額（円）
	AverageRating      *float64 `json:"average_rating"`       // 平均評価（nullの可能性あり）
	BooksThisMonth     int      `json:"books_this_month"`     // 今月購入した書籍数
	CompletedThisMonth int      `json:"completed_this_month"` // 今月読了した書籍数
}

// bookUsecase はBookUsecaseインターフェースの実装
// リポジトリとバリデータを保持して、ビジネスロジックを実行
type bookUsecase struct {
	bookRepo  repository.BookRepository // データアクセス用のリポジトリ
	validator *validator.Validate       // 入力データ検証用のバリデータ
}

// NewBookUsecase は新しいBookUsecaseを作成する関数
// コンストラクタ関数：依存関係を注入してインスタンスを作成
func NewBookUsecase(bookRepo repository.BookRepository) BookUsecase {
	return &bookUsecase{
		bookRepo:  bookRepo,        // リポジトリを設定
		validator: validator.New(), // バリデータの新しいインスタンスを作成
	}
}

// CreateBook は新しい書籍を作成する関数
// ビジネスルール：入力データの検証、購入日のチェックなど
func (u *bookUsecase) CreateBook(req *model.CreateBookRequest) (*model.Book, error) {
	// バリデーション：入力データが正しいかをチェック
	// validator.Struct()：構造体のタグ（requiredなど）をチェック
	if err := u.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("入力データが無効です: %w", err)
	}

	// ビジネスルール：購入日が未来でないことを確認
	// time.Now().After()：指定した時刻より後かどうかを判定
	if req.PurchaseDate.After(time.Now()) {
		return nil, fmt.Errorf("購入日は現在以前の日付を指定してください")
	}

	// 検証が成功したらリポジトリに作成を依頼
	return u.bookRepo.Create(req)
}

// GetBook は指定されたIDの書籍を取得する関数
// ビジネスルール：IDの有効性をチェック（正の整数のみ有効）
func (u *bookUsecase) GetBook(id int) (*model.Book, error) {
	// IDの有効性チェック：0以下はNG（データベースのIDは通常1から始まる）
	if id <= 0 {
		return nil, fmt.Errorf("無効な書籍IDです: %d", id)
	}

	// 検証が成功したらリポジトリに取得を依頼
	return u.bookRepo.GetByID(id)
}

// ListBooks は書籍一覧を取得する関数（ページネーション対応）
// ページネーション：大量のデータをページ単位で分割して表示する機能
func (u *bookUsecase) ListBooks(filter *model.BookFilter, page, limit int) ([]*model.Book, int, error) {
	// ページ番号の正規化：1未満の場合は1に修正
	if page < 1 {
		page = 1
	}
	// 1ページあたりの件数の正規化：1-100の範囲内、デフォルトは20件
	if limit < 1 || limit > 100 {
		limit = 20 // デフォルト値（適切なパフォーマンスを保つ）
	}

	// offset：スキップする件数を計算（ページ番号から算出）
	// 例：2ページ目、】1ページに20件なら offset = (2-1) * 20 = 20
	offset := (page - 1) * limit

	// リポジトリから書籍一覧を取得
	books, err := u.bookRepo.List(filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// 総件数を取得（ページネーション表示用）
	total, err := u.bookRepo.Count(filter)
	if err != nil {
		return nil, 0, err
	}

	// 書籍リストと総件数を返す
	return books, total, nil
}

// UpdateBook は書籍情報を更新する関数
// ビジネスルール：IDの有効性、存在確認、評価の範囲チェック
func (u *bookUsecase) UpdateBook(id int, req *model.UpdateBookRequest) (*model.Book, error) {
	// IDの有効性チェック
	if id <= 0 {
		return nil, fmt.Errorf("無効な書籍IDです: %d", id)
	}

	// 既存の書籍が存在するか確認（存在しないと更新できない）
	_, err := u.bookRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// ビジネスルール：評価が1-5の範囲内かチェック
	// nilチェックが必要（評価が設定されていない場合もある）
	if req.Rating != nil && (*req.Rating < 1 || *req.Rating > 5) {
		return nil, fmt.Errorf("評価は1-5の範囲で入力してください: %d", *req.Rating)
	}

	// 検証が成功したらリポジトリに更新を依頼
	return u.bookRepo.Update(id, req)
}

// DeleteBook は書籍を削除する関数
// ビジネスルール：IDの有効性、存在確認を前もって削除実行
func (u *bookUsecase) DeleteBook(id int) error {
	// IDの有効性チェック
	if id <= 0 {
		return fmt.Errorf("無効な書籍IDです: %d", id)
	}

	// 既存の書籍が存在するか確認（存在しないものは削除できない）
	_, err := u.bookRepo.GetByID(id)
	if err != nil {
		return err
	}

	// 検証が成功したらリポジトリに削除を依頼
	return u.bookRepo.Delete(id)
}

// StartReading は読書を開始する関数
// ビジネスルール：未読または中断状態の書籍のみ読書開始可能
func (u *bookUsecase) StartReading(id int) (*model.Book, error) {
	// IDの有効性チェック
	if id <= 0 {
		return nil, fmt.Errorf("無効な書籍IDです: %d", id)
	}

	// 現在の書籍情報を取得して、ステータスを確認
	book, err := u.bookRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// ビジネスルール：既に読書中または読了している場合はエラー
	if book.Status == model.StatusReading {
		return nil, fmt.Errorf("この書籍は既に読書中です")
	}
	if book.Status == model.StatusCompleted {
		return nil, fmt.Errorf("この書籍は既に読了済みです")
	}

	// 読書開始処理：ステータスを「読書中」に変更、開始日を記録
	status := model.StatusReading
	now := time.Now()  // 現在時刻を取得
	updateReq := &model.UpdateBookRequest{
		Status:        &status, // ステータスを読書中に設定
		StartReadDate: &now,    // 読書開始日を現在時刻に設定
	}

	// リポジトリに更新を依頼
	return u.bookRepo.Update(id, updateReq)
}

// FinishReading は読書を完了する関数
// ビジネスルール：読書中の書籍のみ完了可能、評価は任意で、1-5の範囲
func (u *bookUsecase) FinishReading(id int, rating *int) (*model.Book, error) {
	// IDの有効性チェック
	if id <= 0 {
		return nil, fmt.Errorf("無効な書籍IDです: %d", id)
	}

	// 現在の書籍情報を取得して、ステータスを確認
	book, err := u.bookRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// ビジネスルール：読書中でない場合はエラー
	if book.Status != model.StatusReading {
		return nil, fmt.Errorf("この書籍は読書中ではありません")
	}

	// 評価のバリデーション：設定されている場合は1-5の範囲内かチェック
	if rating != nil && (*rating < 1 || *rating > 5) {
		return nil, fmt.Errorf("評価は1-5の範囲で入力してください: %d", *rating)
	}

	// 読書完了処理：ステータスを「読了」に変更、終了日と評価を記録
	status := model.StatusCompleted
	now := time.Now()  // 現在時刻を取得
	updateReq := &model.UpdateBookRequest{
		Status:      &status, // ステータスを読了に設定
		EndReadDate: &now,    // 読書終了日を現在時刻に設定
		Rating:      rating,  // 評価を設定（nilの場合は評価なし）
	}

	// リポジトリに更新を依頼
	return u.bookRepo.Update(id, updateReq)
}

// GetStatistics は書籍の統計情報を取得する関数
// 複雑な集計処理：全書籍データを取得して様々な統計値を計算
func (u *bookUsecase) GetStatistics() (*BookStatistics, error) {
	// 空の統計情報構造体を作成（これから各フィールドに値を設定していく）
	stats := &BookStatistics{}

	// 全書籍数を取得
	// Count(nil)：フィルター条件なし（全件）でカウント
	total, err := u.bookRepo.Count(nil)
	if err != nil {
		return nil, fmt.Errorf("統計情報の取得に失敗しました: %w", err)
	}
	stats.TotalBooks = total  // 総書籍数を設定

	// ステータス別の書籍数を効率的に取得
	// map：キーと値のペア（辞書）、ここではステータスと設定先のポインタを対応付け
	statusCounts := map[model.ReadingStatus]*int{
		model.StatusNotStarted: &stats.NotStartedBooks, // 未読数の設定先
		model.StatusReading:    &stats.ReadingBooks,    // 読書中数の設定先
		model.StatusCompleted:  &stats.CompletedBooks,  // 読了数の設定先
		model.StatusDropped:    &stats.DroppedBooks,    // 中断数の設定先
	}

	// 各ステータスごとにループして件数を取得
	// range：mapやスライスの要素を順番に処理するループ
	for status, countPtr := range statusCounts {
		// 特定のステータスのみを対象とするフィルターを作成
		filter := &model.BookFilter{Status: &status}
		count, err := u.bookRepo.Count(filter)
		if err != nil {
			return nil, fmt.Errorf("ステータス別統計の取得に失敗しました: %w", err)
		}
		// *countPtr：ポインタの指す先に値を代入（stats構造体の該当フィールドに設定）
		*countPtr = count
	}

	// 全書籍を取得して金額や評価を計算
	// List(nil, 0, 0)：フィルターなし、制限なしで全書籍を取得
	allBooks, err := u.bookRepo.List(nil, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("書籍一覧の取得に失敗しました: %w", err)
	}

	// 集計用の変数を初期化
	totalSpent := 0         // 総支出額の累計
	ratingSum := 0          // 評価の合計値（平均計算用）
	ratingCount := 0        // 評価された書籍の数（平均計算用）
	booksThisMonth := 0     // 今月購入した書籍数
	completedThisMonth := 0 // 今月完了した書籍数
	
	// 時間計算：今月の開始日時を計算
	now := time.Now()  // 現在日時を取得
	// time.Date()：指定した年月日時分秒の時刻を作成
	// now.Year(), now.Month(), 1：今年今月の1日を指定
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 全書籍をループして各種統計を計算
	// range：スライスの各要素を順番に処理（_はインデックスを無視）
	for _, book := range allBooks {
		// 合計支出額の累計
		// +=：現在の値に加算して代入
		totalSpent += book.PurchasePrice

		// 評価の平均値計算のための合計値とカウント
		// nil チェック：評価が設定されている書籍のみ対象
		if book.Rating != nil {
			ratingSum += *book.Rating  // 評価の合計に加算
			ratingCount++              // 評価された書籍数をカウント
		}

		// 今月購入した書籍数をカウント
		// After()：指定時刻より後かチェック、Equal()：同じ時刻かチェック
		if book.PurchaseDate.After(thisMonthStart) || book.PurchaseDate.Equal(thisMonthStart) {
			booksThisMonth++
		}

		// 今月完了した書籍数をカウント
		// 複数条件：終了日が設定されている && 今月内 && ステータスが完了
		if book.EndReadDate != nil && 
		   (book.EndReadDate.After(thisMonthStart) || book.EndReadDate.Equal(thisMonthStart)) &&
		   book.Status == model.StatusCompleted {
			completedThisMonth++
		}
	}

	// 計算結果を統計情報構造体に設定
	stats.TotalSpent = totalSpent                   // 総支出額
	stats.BooksThisMonth = booksThisMonth           // 今月購入数
	stats.CompletedThisMonth = completedThisMonth   // 今月完了数

	// 平均評価の計算（評価された書籍がある場合のみ）
	if ratingCount > 0 {
		// 型変換：int を float64 に変換して小数点付きの平均値を計算
		avg := float64(ratingSum) / float64(ratingCount)
		stats.AverageRating = &avg  // ポインタで設定（nil の可能性を表現）
	}
	// ratingCount が 0 の場合、AverageRating は nil のまま（評価なし）

	// 完成した統計情報を返す
	return stats, nil
}
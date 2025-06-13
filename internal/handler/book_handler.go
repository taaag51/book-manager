// handlerパッケージ：HTTPリクエストを処理するファイル
// Handler層は「プレゼンテーション層」とも呼ばれ、Webブラウザからのリクエストを受け取る
package handler

// import：他のパッケージ（機能）を使うための宣言
import (
	"encoding/json"                      // JSONデータのエンコード（変換）・デコード（解析）
	"net/http"                          // HTTPサーバー機能（リクエスト・レスポンス処理）
	"strconv"                           // 文字列と数値の変換（"123" → 123など）

	"book-manager/internal/model"        // 自作のデータ構造定義
	"book-manager/internal/usecase"      // 自作のビジネスロジック層
	"github.com/gorilla/mux"             // URLルーティングライブラリ（URLと処理の対応付け）
)

// BookHandler は書籍関連のHTTPリクエストを処理する構造体
// HTTPリクエスト：Webブラウザからサーバーへのデータ送信（GET、POSTなど）
type BookHandler struct {
	bookUsecase usecase.BookUsecase // ビジネスロジック処理用のユースケース
}

// NewBookHandler は新しいBookHandlerを作成する関数
// コンストラクタ関数：依存関係を注入してインスタンスを作成
func NewBookHandler(bookUsecase usecase.BookUsecase) *BookHandler {
	return &BookHandler{bookUsecase: bookUsecase} // ユースケースを設定したハンドラを返す
}

// ErrorResponse はエラーレスポンスの構造体
// エラー発生時にクライアントに返すJSONデータの形式
type ErrorResponse struct {
	Error   string `json:"error"`   // エラーの種類（ユーザー向けメッセージ）
	Message string `json:"message"` // 詳細なエラー内容（デバッグ用）
}

// SuccessResponse は成功レスポンスの構造体
// 処理成功時にクライアントに返すJSONデータの形式
type SuccessResponse struct {
	Message string      `json:"message"`          // 成功メッセージ
	Data    interface{} `json:"data,omitempty"`   // 実際のデータ（interface{}は任意の型を表す）
}

// ListBooksResponse は書籍一覧レスポンスの構造体
// ページング情報を含む書籍一覧を返すための専用構造体
type ListBooksResponse struct {
	Books      []*model.Book `json:"books"`       // 書籍データの配列
	Total      int           `json:"total"`       // 総件数（全書籍数）
	Page       int           `json:"page"`        // 現在のページ番号
	Limit      int           `json:"limit"`       // 1ページあたりの件数
	TotalPages int           `json:"total_pages"` // 総ページ数
}

// CreateBook は新しい書籍を作成するHTTPハンドラ関数
// POST /api/v1/books のリクエストを処理
// w: レスポンス書き込み用、r: リクエスト情報読み取り用
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	// リクエストボディからJSONデータを解析して構造体に変換
	var req model.CreateBookRequest
	// json.NewDecoder(r.Body).Decode()：HTTPリクエストのJSONをGoの構造体に変換
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// パースエラーの場合は400 Bad Requestでエラーレスポンスを返す
		h.sendErrorResponse(w, http.StatusBadRequest, "リクエストの解析に失敗しました", err)
		return
	}

	// ユースケースでビジネスロジックを実行（バリデーション、データ保存）
	book, err := h.bookUsecase.CreateBook(&req)
	if err != nil {
		// ビジネスロジックエラーの場合は400 Bad Requestでエラーレスポンスを返す
		h.sendErrorResponse(w, http.StatusBadRequest, "書籍の作成に失敗しました", err)
		return
	}

	// 成功時は201 Createdで作成された書籍データを返す
	h.sendSuccessResponse(w, http.StatusCreated, "書籍が正常に作成されました", book)
}

// GetBook は指定されたIDの書籍を取得するHTTPハンドラ関数
// GET /api/v1/books/{id} のリクエストを処理
func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	// URLパスからパラメータを取得
	// mux.Vars()：Gorilla Muxで定義したURLパラメータを取得
	vars := mux.Vars(r)
	idStr := vars["id"]  // "{id}"で定義した部分の値を取得

	// 文字列のIDを数値に変換
	// strconv.Atoi()：文字列を整数に変換（"123" → 123）
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// 数値変換エラーの場合は400 Bad Request
		h.sendErrorResponse(w, http.StatusBadRequest, "無効な書籍IDです", err)
		return
	}

	// ユースケースで書籍情報を取得
	book, err := h.bookUsecase.GetBook(id)
	if err != nil {
		// 書籍が見つからない場合は404 Not Found
		h.sendErrorResponse(w, http.StatusNotFound, "書籍が見つかりません", err)
		return
	}

	// 成功時は200 OKで書籍データを返す（メッセージは空）
	h.sendSuccessResponse(w, http.StatusOK, "", book)
}

// ListBooks は書籍一覧を取得するHTTPハンドラ関数
// GET /api/v1/books?page=1&limit=20&status=reading などのリクエストを処理
func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	// URLクエリパラメータを取得
	// r.URL.Query()：?page=1&limit=20 のようなクエリパラメータを取得
	query := r.URL.Query()

	// ページネーションパラメータの解析とバリデーション
	// _：エラーを無視（変換失敗時は0が返る）
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1  // ページ番号は最低1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20  // デフォルトは20件、最大60100件まで
	}

	// フィルター条件を構築（検索、絞り込み条件）
	filter := &model.BookFilter{}

	// 各種フィルターパラメータをチェックして設定
	// パラメータが空でない場合のみフィルターに設定
	if status := query.Get("status"); status != "" {
		// 文字列をReadingStatus型に変換
		readingStatus := model.ReadingStatus(status)
		filter.Status = &readingStatus  // ポインタで設定
	}

	if author := query.Get("author"); author != "" {
		filter.Author = &author  // 著者名で絞り込み
	}

	if publisher := query.Get("publisher"); publisher != "" {
		filter.Publisher = &publisher  // 出版社で絞り込み
	}

	if tag := query.Get("tag"); tag != "" {
		filter.Tag = &tag  // タグで絞り込み
	}

	if search := query.Get("search"); search != "" {
		filter.Search = &search  // タイトル・著者の部分一致検索
	}

	// 評価パラメータは数値バリデーションが必要
	if ratingStr := query.Get("rating"); ratingStr != "" {
		// 数値変換と範囲チェック（1-5の範囲内のみ有効）
		if rating, err := strconv.Atoi(ratingStr); err == nil && rating >= 1 && rating <= 5 {
			filter.Rating = &rating
		}
	}

	// ユースケースで書籍一覧を取得（フィルター、ページング付き）
	books, total, err := h.bookUsecase.ListBooks(filter, page, limit)
	if err != nil {
		// サーバー内部エラーの場合は500 Internal Server Error
		h.sendErrorResponse(w, http.StatusInternalServerError, "書籍一覧の取得に失敗しました", err)
		return
	}

	// 総ページ数を計算（割り算の切り上げ）
	// (total + limit - 1) / limit：切り上げ除算のテクニック
	totalPages := (total + limit - 1) / limit
	// ページング情報を含むレスポンスを構築
	response := ListBooksResponse{
		Books:      books,      // 書籍データの配列
		Total:      total,      // 総件数
		Page:       page,       // 現在ページ
		Limit:      limit,      // 1ページあたりの件数
		TotalPages: totalPages, // 総ページ数
	}

	// 成功時は200 OKでページング情報付き一覧を返す
	h.sendSuccessResponse(w, http.StatusOK, "", response)
}

// UpdateBook は書籍情報を更新するHTTPハンドラ関数
// PUT /api/v1/books/{id} のリクエストを処理
func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	// URLパスから書籍IDを取得
	vars := mux.Vars(r)
	idStr := vars["id"]

	// 文字列IDを数値に変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "無効な書籍IDです", err)
		return
	}

	// リクエストボディから更新データを解析
	var req model.UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "リクエストの解析に失敗しました", err)
		return
	}

	// ユースケースで書籍情報を更新
	book, err := h.bookUsecase.UpdateBook(id, &req)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "書籍の更新に失敗しました", err)
		return
	}

	// 成功時は200 OKで更新後の書籍データを返す
	h.sendSuccessResponse(w, http.StatusOK, "書籍が正常に更新されました", book)
}

// DeleteBook は書籍を削除するHTTPハンドラ関数
// DELETE /api/v1/books/{id} のリクエストを処理
func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	// URLパスから書籍IDを取得
	vars := mux.Vars(r)
	idStr := vars["id"]

	// 文字列IDを数値に変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "無効な書籍IDです", err)
		return
	}

	// ユースケースで書籍を削除
	if err := h.bookUsecase.DeleteBook(id); err != nil {
		// 書籍が見つからないまたは削除失敗の場合は404 Not Found
		h.sendErrorResponse(w, http.StatusNotFound, "書籍の削除に失敗しました", err)
		return
	}

	// 成功時は200 OKでメッセージを返す（データはnil）
	h.sendSuccessResponse(w, http.StatusOK, "書籍が正常に削除されました", nil)
}

// StartReading は読書を開始するHTTPハンドラ関数
// POST /api/v1/books/{id}/start-reading のリクエストを処理
func (h *BookHandler) StartReading(w http.ResponseWriter, r *http.Request) {
	// URLパスから書籍IDを取得
	vars := mux.Vars(r)
	idStr := vars["id"]

	// 文字列IDを数値に変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "無効な書籍IDです", err)
		return
	}

	// ユースケースで読書を開始（ステータスを読書中に変更）
	book, err := h.bookUsecase.StartReading(id)
	if err != nil {
		// ビジネスルールエラー（既に読書中など）の場合は400 Bad Request
		h.sendErrorResponse(w, http.StatusBadRequest, "読書開始に失敗しました", err)
		return
	}

	// 成功時は200 OKで更新後の書籍データを返す
	h.sendSuccessResponse(w, http.StatusOK, "読書を開始しました", book)
}

// FinishReading は読書を完了するHTTPハンドラ関数
// POST /api/v1/books/{id}/finish-reading のリクエストを処理
func (h *BookHandler) FinishReading(w http.ResponseWriter, r *http.Request) {
	// URLパスから書籍IDを取得
	vars := mux.Vars(r)
	idStr := vars["id"]

	// 文字列IDを数値に変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "無効な書籍IDです", err)
		return
	}

	// リクエストボディから評価を取得（オプション）
	// 無名構造体：一時的なデータ受け取り用
	var reqBody struct {
		Rating *int `json:"rating"`  // 評価（1-5点、任意）
	}

	// ContentLength > 0：リクエストボディがある場合のみ解析
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			h.sendErrorResponse(w, http.StatusBadRequest, "リクエストの解析に失敗しました", err)
			return
		}
	}

	// ユースケースで読書を完了（ステータスを完了に変更、評価設定）
	book, err := h.bookUsecase.FinishReading(id, reqBody.Rating)
	if err != nil {
		// ビジネスルールエラー（読書中でないなど）の場合は400 Bad Request
		h.sendErrorResponse(w, http.StatusBadRequest, "読書完了に失敗しました", err)
		return
	}

	// 成功時は200 OKで更新後の書籍データを返す
	h.sendSuccessResponse(w, http.StatusOK, "読書を完了しました", book)
}

// GetStatistics は書籍の統計情報を取得するHTTPハンドラ関数
// GET /api/v1/statistics のリクエストを処理
func (h *BookHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	// ユースケースで統計情報を取得（合計金額、平均評価など）
	stats, err := h.bookUsecase.GetStatistics()
	if err != nil {
		// サーバー内部エラーの場合は500 Internal Server Error
		h.sendErrorResponse(w, http.StatusInternalServerError, "統計情報の取得に失敗しました", err)
		return
	}

	// 成功時は200 OKで統計データを返す
	h.sendSuccessResponse(w, http.StatusOK, "", stats)
}

// Health はヘルスチェック用のHTTPハンドラ関数
// GET /api/v1/health のリクエストを処理（サーバーの動作状態を確認）
func (h *BookHandler) Health(w http.ResponseWriter, r *http.Request) {
	// 常に200 OKでサービスの動作状態を返す（モニタリング用）
	h.sendSuccessResponse(w, http.StatusOK, "サービスは正常に動作しています", map[string]string{
		"status": "healthy",      // サービスの状態
		"service": "book-manager", // サービス名
	})
}

// sendErrorResponse はエラーレスポンスを送信するヘルパー関数
// 共通のエラー処理をまとめて、コードの重複を防ぐ
func (h *BookHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	// HTTPレスポンスヘッダーを設定（JSON形式で返すことを明示）
	w.Header().Set("Content-Type", "application/json")
	// HTTPステータスコードを設定（400, 404, 500など）
	w.WriteHeader(statusCode)

	// エラーレスポンス構造体を作成
	response := ErrorResponse{
		Error:   message,    // ユーザー向けエラーメッセージ
		Message: err.Error(), // 詳細なエラー内容（デバッグ用）
	}

	// JSON形式でレスポンスを送信
	json.NewEncoder(w).Encode(response)
}

// sendSuccessResponse は成功レスポンスを送信するヘルパー関数
// 共通の成功処理をまとめて、コードの重複を防ぐ
func (h *BookHandler) sendSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	// HTTPレスポンスヘッダーを設定（JSON形式で返すことを明示）
	w.Header().Set("Content-Type", "application/json")
	// HTTPステータスコードを設定（200, 201など）
	w.WriteHeader(statusCode)

	// 成功レスポンス構造体を作成
	response := SuccessResponse{
		Message: message, // 成功メッセージ
		Data:    data,    // 実際のデータ（interface{}は任意の型を受け入れる）
	}

	// JSON形式でレスポンスを送信
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes はHTTPルートを登録する関数
// URLパスとHTTPメソッドを組み合わせて、処理関数を割り当てる
func (h *BookHandler) RegisterRoutes(router *mux.Router) {
	// 書籍CRUD操作（Create, Read, Update, Delete）
	// CRUD：データの作成・取得・更新・削除の基本操作
	router.HandleFunc("/books", h.CreateBook).Methods("POST")                        // 書籍作成
	router.HandleFunc("/books", h.ListBooks).Methods("GET")                         // 書籍一覧取得
	router.HandleFunc("/books/{id:[0-9]+}", h.GetBook).Methods("GET")               // 書籍1件取得
	router.HandleFunc("/books/{id:[0-9]+}", h.UpdateBook).Methods("PUT")            // 書籍更新
	router.HandleFunc("/books/{id:[0-9]+}", h.DeleteBook).Methods("DELETE")         // 書籍削除
	// {id:[0-9]+}：URLパラメータで数字のみIDとして受け入れる

	// 読書管理操作（ビジネスロジック固有の操作）
	router.HandleFunc("/books/{id:[0-9]+}/start-reading", h.StartReading).Methods("POST")   // 読書開始
	router.HandleFunc("/books/{id:[0-9]+}/finish-reading", h.FinishReading).Methods("POST") // 読書完了

	// 統計情報取得
	router.HandleFunc("/statistics", h.GetStatistics).Methods("GET")  // 書籍統計情報

	// ヘルスチェック（サーバーの動作確認用）
	router.HandleFunc("/health", h.Health).Methods("GET")             // サービスの動作状態確認
}
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"book-manager/internal/model"
	"book-manager/internal/usecase"
	"github.com/gorilla/mux"
)

// BookHandler は書籍関連のHTTPリクエストを処理する
type BookHandler struct {
	bookUsecase usecase.BookUsecase
}

// NewBookHandler は新しいBookHandlerを作成する
func NewBookHandler(bookUsecase usecase.BookUsecase) *BookHandler {
	return &BookHandler{bookUsecase: bookUsecase}
}

// ErrorResponse はエラーレスポンスの構造体
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse は成功レスポンスの構造体
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ListBooksResponse は書籍一覧レスポンスの構造体
type ListBooksResponse struct {
	Books      []*model.Book `json:"books"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// CreateBook は新しい書籍を作成する
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "リクエストの解析に失敗しました", err)
		return
	}

	book, err := h.bookUsecase.CreateBook(&req)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "書籍の作成に失敗しました", err)
		return
	}

	h.sendSuccessResponse(w, http.StatusCreated, "書籍が正常に作成されました", book)
}

// GetBook は指定されたIDの書籍を取得する
func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "無効な書籍IDです", err)
		return
	}

	book, err := h.bookUsecase.GetBook(id)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "書籍が見つかりません", err)
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, "", book)
}

// ListBooks は書籍一覧を取得する
func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// ページネーションパラメータ
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// フィルターパラメータ
	filter := &model.BookFilter{}

	if status := query.Get("status"); status != "" {
		readingStatus := model.ReadingStatus(status)
		filter.Status = &readingStatus
	}

	if author := query.Get("author"); author != "" {
		filter.Author = &author
	}

	if publisher := query.Get("publisher"); publisher != "" {
		filter.Publisher = &publisher
	}

	if tag := query.Get("tag"); tag != "" {
		filter.Tag = &tag
	}

	if search := query.Get("search"); search != "" {
		filter.Search = &search
	}

	if ratingStr := query.Get("rating"); ratingStr != "" {
		if rating, err := strconv.Atoi(ratingStr); err == nil && rating >= 1 && rating <= 5 {
			filter.Rating = &rating
		}
	}

	books, total, err := h.bookUsecase.ListBooks(filter, page, limit)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "書籍一覧の取得に失敗しました", err)
		return
	}

	totalPages := (total + limit - 1) / limit
	response := ListBooksResponse{
		Books:      books,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	h.sendSuccessResponse(w, http.StatusOK, "", response)
}

// UpdateBook は書籍情報を更新する
func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "無効な書籍IDです", err)
		return
	}

	var req model.UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "リクエストの解析に失敗しました", err)
		return
	}

	book, err := h.bookUsecase.UpdateBook(id, &req)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "書籍の更新に失敗しました", err)
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, "書籍が正常に更新されました", book)
}

// DeleteBook は書籍を削除する
func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "無効な書籍IDです", err)
		return
	}

	if err := h.bookUsecase.DeleteBook(id); err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "書籍の削除に失敗しました", err)
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, "書籍が正常に削除されました", nil)
}

// StartReading は読書を開始する
func (h *BookHandler) StartReading(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "無効な書籍IDです", err)
		return
	}

	book, err := h.bookUsecase.StartReading(id)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "読書開始に失敗しました", err)
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, "読書を開始しました", book)
}

// FinishReading は読書を完了する
func (h *BookHandler) FinishReading(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "無効な書籍IDです", err)
		return
	}

	// リクエストボディから評価を取得（オプション）
	var reqBody struct {
		Rating *int `json:"rating"`
	}

	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			h.sendErrorResponse(w, http.StatusBadRequest, "リクエストの解析に失敗しました", err)
			return
		}
	}

	book, err := h.bookUsecase.FinishReading(id, reqBody.Rating)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "読書完了に失敗しました", err)
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, "読書を完了しました", book)
}

// GetStatistics は書籍の統計情報を取得する
func (h *BookHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.bookUsecase.GetStatistics()
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, "統計情報の取得に失敗しました", err)
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, "", stats)
}

// Health はヘルスチェック用のエンドポイント
func (h *BookHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.sendSuccessResponse(w, http.StatusOK, "サービスは正常に動作しています", map[string]string{
		"status": "healthy",
		"service": "book-manager",
	})
}

// sendErrorResponse はエラーレスポンスを送信する
func (h *BookHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   message,
		Message: err.Error(),
	}

	json.NewEncoder(w).Encode(response)
}

// sendSuccessResponse は成功レスポンスを送信する
func (h *BookHandler) sendSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := SuccessResponse{
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes はルートを登録する
func (h *BookHandler) RegisterRoutes(router *mux.Router) {
	// 書籍CRUD操作
	router.HandleFunc("/books", h.CreateBook).Methods("POST")
	router.HandleFunc("/books", h.ListBooks).Methods("GET")
	router.HandleFunc("/books/{id:[0-9]+}", h.GetBook).Methods("GET")
	router.HandleFunc("/books/{id:[0-9]+}", h.UpdateBook).Methods("PUT")
	router.HandleFunc("/books/{id:[0-9]+}", h.DeleteBook).Methods("DELETE")

	// 読書管理操作
	router.HandleFunc("/books/{id:[0-9]+}/start-reading", h.StartReading).Methods("POST")
	router.HandleFunc("/books/{id:[0-9]+}/finish-reading", h.FinishReading).Methods("POST")

	// 統計情報
	router.HandleFunc("/statistics", h.GetStatistics).Methods("GET")

	// ヘルスチェック
	router.HandleFunc("/health", h.Health).Methods("GET")
}
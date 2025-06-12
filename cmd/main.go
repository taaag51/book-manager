package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"book-manager/internal/database"
	"book-manager/internal/handler"
	"book-manager/internal/repository"
	"book-manager/internal/usecase"
	"github.com/gorilla/mux"
)

const (
	defaultPort     = "8080"
	defaultDBPath   = "./books.db"
	shutdownTimeout = 30 * time.Second
)

func main() {
	// 環境変数から設定を取得
	port := getEnv("PORT", defaultPort)
	dbPath := getEnv("DB_PATH", defaultDBPath)

	// データベース接続の初期化
	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Fatalf("データベースの初期化に失敗しました: %v", err)
	}
	defer db.Close()

	// マイグレーションの実行
	if err := db.Migrate(); err != nil {
		log.Fatalf("マイグレーションに失敗しました: %v", err)
	}

	// 依存関係の注入
	bookRepo := repository.NewBookRepository(db)
	bookUsecase := usecase.NewBookUsecase(bookRepo)
	bookHandler := handler.NewBookHandler(bookUsecase)

	// ルーターの設定
	router := mux.NewRouter()
	
	// CORS設定
	router.Use(corsMiddleware)
	// ログ出力ミドルウェア
	router.Use(loggingMiddleware)

	// API ルートの登録
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	bookHandler.RegisterRoutes(apiRouter)

	// 静的ファイル配信（CSS、JS、画像）
	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("./web/css/")))).Methods("GET")
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./web/js/")))).Methods("GET")
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("./web/images/")))).Methods("GET")
	
	// ルートパスとindex.htmlの配信
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	}).Methods("GET")

	// HTTPサーバーの設定
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// サーバーの開始
	go func() {
		log.Printf("書籍管理サーバーを開始します。ポート: %s", port)
		log.Printf("WebUI: http://localhost:%s", port)
		log.Printf("API エンドポイント: http://localhost:%s/api/v1", port)
		log.Printf("ヘルスチェック: http://localhost:%s/api/v1/health", port)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("サーバーの開始に失敗しました: %v", err)
		}
	}()

	// グレースフルシャットダウンの設定
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("サーバーをシャットダウンしています...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("サーバーのシャットダウンに失敗しました: %v", err)
	}

	log.Println("サーバーが正常にシャットダウンされました")
}

// getEnv は環境変数を取得し、存在しない場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// corsMiddleware はCORSヘッダーを設定するミドルウェア
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware はリクエストをログ出力するミドルウェア
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// レスポンスライターをラップしてステータスコードを取得
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf(
			"%s %s %d %v %s",
			r.Method,
			r.RequestURI,
			lrw.statusCode,
			duration,
			r.UserAgent(),
		)
	})
}

// loggingResponseWriter はレスポンスライターのラッパー
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
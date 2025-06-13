// これはGoプログラムのエントリーポイント（開始地点）です
// package mainと書くことで、このファイルが実行可能なプログラムになります
package main

// importは他のパッケージ（機能）を使うための宣言です
// 例：log → ログ出力、net/http → Webサーバー機能
import (
	"context"                               // プログラムのキャンセル処理
	"log"                                   // ログ（記録）を出力する
	"net/http"                              // Webサーバーを作る
	"os"                                    // OS（オペレーティングシステム）とやり取り
	"os/signal"                             // プログラム終了信号をキャッチ
	"syscall"                               // システムコール（OS機能）
	"time"                                  // 時間関連の処理

	"book-manager/internal/database"        // データベース関連の機能
	"book-manager/internal/handler"         // HTTPリクエストを処理する機能
	"book-manager/internal/repository"      // データの保存・取得機能
	"book-manager/internal/usecase"         // ビジネスロジック（業務処理）
	"github.com/gorilla/mux"                // URLルーティング（アドレス振り分け）
)

// constは定数（変わらない値）を定義します
const (
	defaultPort     = "8080"              // デフォルトのポート番号（Webサーバーが使う番号）
	defaultDBPath   = "./books.db"        // データベースファイルの保存場所
	shutdownTimeout = 30 * time.Second    // サーバー停止時の待機時間（30秒）
)

// main関数：プログラムが開始される場所です
// Goプログラムは必ずmain関数から実行されます
func main() {
	// 環境変数から設定を取得
	// 環境変数とは：OSに設定されている変数（設定値）
	// もし環境変数が設定されていなければ、デフォルト値を使用
	port := getEnv("PORT", defaultPort)
	dbPath := getEnv("DB_PATH", defaultDBPath)

	// データベース接続の初期化
	// データベースとは：データを保存する場所
	// NewDB()でデータベースに接続する準備をします
	db, err := database.NewDB(dbPath)
	if err != nil {
		// エラーが発生した場合、プログラムを終了
		log.Fatalf("データベースの初期化に失敗しました: %v", err)
	}
	// defer：この関数が終了する時に実行される処理
	// プログラム終了時にデータベース接続を閉じる
	defer db.Close()

	// マイグレーションの実行
	// マイグレーション：データベースにテーブル（表）を作成する処理
	if err := db.Migrate(); err != nil {
		log.Fatalf("マイグレーションに失敗しました: %v", err)
	}

	// 依存関係の注入（Dependency Injection）
	// 各層（Repository、UseCase、Handler）を作成し、連携させる
	// Repository：データの保存・取得を担当
	// UseCase：業務ロジック（書籍の管理方法）を担当
	// Handler：Webリクエストの処理を担当
	bookRepo := repository.NewBookRepository(db)        // データアクセス層
	bookUsecase := usecase.NewBookUsecase(bookRepo)     // ビジネスロジック層
	bookHandler := handler.NewBookHandler(bookUsecase)  // プレゼンテーション層

	// ルーターの設定
	// ルーターとは：URLに応じてどの処理を実行するかを決める仕組み
	// 例：/api/v1/books → 書籍一覧を表示
	router := mux.NewRouter()
	
	// CORS設定
	// CORS：ブラウザから別のドメインのAPIを呼び出すための設定
	router.Use(corsMiddleware)
	
	// ログ出力ミドルウェア
	// ミドルウェア：リクエストの前後で共通処理を行う仕組み
	// アクセスログ（誰がいつアクセスしたか）を記録
	router.Use(loggingMiddleware)

	// API ルートの登録
	// /api/v1 で始まるURLをAPIとして扱う
	// 例：/api/v1/books、/api/v1/statistics など
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	bookHandler.RegisterRoutes(apiRouter)

	// 静的ファイル配信（CSS、JS、画像）
	// 静的ファイル：変更されないファイル（CSSやJavaScriptなど）
	// /css/style.css → ./web/css/style.css を返す
	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("./web/css/")))).Methods("GET")
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("./web/js/")))).Methods("GET")
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("./web/images/")))).Methods("GET")
	
	// ルートパス（トップページ）の設定
	// http://localhost:8080/ にアクセスした時に表示するページ
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// index.htmlファイルを返す
		http.ServeFile(w, r, "./web/index.html")
	}).Methods("GET")

	// HTTPサーバーの設定
	// サーバー：Webブラウザからのリクエストを受け取る仕組み
	srv := &http.Server{
		Addr:         ":" + port,             // サーバーが使うポート番号
		Handler:      router,                 // URLルーティング設定
		ReadTimeout:  15 * time.Second,       // リクエスト読み取りのタイムアウト
		WriteTimeout: 15 * time.Second,       // レスポンス書き込みのタイムアウト
		IdleTimeout:  60 * time.Second,       // アイドル状態のタイムアウト
	}

	// サーバーの開始
	// go func()：別の処理として並行実行（ゴルーチン）
	go func() {
		// サーバー開始のメッセージを表示
		log.Printf("書籍管理サーバーを開始します。ポート: %s", port)
		log.Printf("WebUI: http://localhost:%s", port)
		log.Printf("API エンドポイント: http://localhost:%s/api/v1", port)
		log.Printf("ヘルスチェック: http://localhost:%s/api/v1/health", port)
		
		// サーバーを開始（ブロッキング処理）
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("サーバーの開始に失敗しました: %v", err)
		}
	}()

	// グレースフルシャットダウンの設定
	// グレースフル：処理中のリクエストを待ってから終了する方法
	quit := make(chan os.Signal, 1)                          // 終了信号を受け取るチャンネル
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)     // Ctrl+Cなどの終了信号を監視
	<-quit                                                    // 終了信号が来るまで待機

	log.Println("サーバーをシャットダウンしています...")

	// 30秒以内にシャットダウンを完了する
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// サーバーを安全に停止
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("サーバーのシャットダウンに失敗しました: %v", err)
	}

	log.Println("サーバーが正常にシャットダウンされました")
}

// getEnv は環境変数を取得し、存在しない場合はデフォルト値を返す関数
// 環境変数：OS（オペレーティングシステム）に設定された設定値
// 例：PORT=3000 と設定されていれば "3000" を返す
func getEnv(key, defaultValue string) string {
	// os.Getenv()で環境変数を取得
	if value := os.Getenv(key); value != "" {
		return value    // 環境変数が設定されていればその値を返す
	}
	return defaultValue // 設定されていなければデフォルト値を返す
}

// corsMiddleware はCORSヘッダーを設定するミドルウェア関数
// CORS（Cross-Origin Resource Sharing）：
// 異なるドメインからのAPIアクセスを許可する仕組み
func corsMiddleware(next http.Handler) http.Handler {
	// http.HandlerFuncでラップして新しいハンドラーを作成
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// レスポンスヘッダーにCORS設定を追加
		w.Header().Set("Access-Control-Allow-Origin", "*")                                // 全てのドメインからアクセス許可
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // 許可するHTTPメソッド
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")     // 許可するヘッダー

		// OPTIONSリクエスト（プリフライトリクエスト）の処理
		// ブラウザが実際のリクエスト前に送る確認リクエスト
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK) // 200 OKを返す
			return
		}

		// 次のミドルウェアまたはハンドラーに処理を委譲
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware はリクエストをログ出力するミドルウェア関数
// アクセスログ：誰がいつどのページにアクセスしたかを記録
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 処理開始時刻を記録
		start := time.Now()

		// レスポンスライターをラップしてステータスコードを取得
		// ラップ：元の機能を拡張して新しい機能を追加すること
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// 次のハンドラーに処理を委譲
		next.ServeHTTP(lrw, r)

		// 処理にかかった時間を計算
		duration := time.Since(start)
		
		// ログを出力
		// フォーマット：HTTPメソッド URL ステータスコード 実行時間 ユーザーエージェント
		log.Printf(
			"%s %s %d %v %s",
			r.Method,        // HTTPメソッド（GET, POST, PUT, DELETE）
			r.RequestURI,    // リクエストされたURL
			lrw.statusCode,  // HTTPステータスコード（200, 404, 500など）
			duration,        // 処理時間
			r.UserAgent(),   // ブラウザ情報
		)
	})
}

// loggingResponseWriter はレスポンスライターのラッパー構造体
// HTTPレスポンスのステータスコードを記録するために使用
type loggingResponseWriter struct {
	http.ResponseWriter        // 元のResponseWriterを埋め込み
	statusCode          int    // ステータスコードを保存する変数
}

// WriteHeader はHTTPステータスコードを設定する関数
// 元のWriteHeaderを呼び出す前にステータスコードを記録
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code                    // ステータスコードを記録
	lrw.ResponseWriter.WriteHeader(code)     // 元のWriteHeaderを呼び出し
}
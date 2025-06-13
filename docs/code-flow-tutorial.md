# 🔍 コードの流れを追跡しよう！

このドキュメントでは、**ユーザーの操作から結果表示まで**のコードの流れを一気通貫で追跡する方法を説明します。

## 🎯 このドキュメントの目的

- **「ボタンを押したら何が起きるのか？」**が分かるようになる
- **フロントエンド → サーバー → データベース → 結果表示**の全体像を理解
- **実際のコードを順番に読む方法**を身につける

## 🌐 まず：このアプリの構成を理解しよう

```
ブラウザ（フロントエンド） ← HTTP → Go Server（バックエンド） → SQLite（データベース）
     ↓                                    ↓
  HTML/CSS/JS                      Handler → UseCase → Repository
```

**重要**：このプロジェクトは**APIサーバーのみ**です。フロントエンドのHTMLファイルはありません。
そのため、**curlコマンド**や**Postman**などのツールでAPIを直接呼び出します。

## 📊 実例1: 「書籍を1冊追加する」操作を完全に追跡

### Step 1: ユーザーの操作（curlコマンド）

```bash
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Go言語入門",
    "author": "山田太郎",
    "purchase_date": "2024-01-15T00:00:00Z",
    "purchase_price": 3000
  }'
```

**何が起きている？**
- HTTPのPOSTリクエストが送信される
- URLは `/api/v1/books`
- データはJSON形式で送信

### Step 2: Goサーバーがリクエストを受信

**📂 ファイル**: `cmd/main.go`

```go
// 77行目あたり - サーバー起動部分
log.Printf("サーバーを開始します (ポート: %s)", port)
if err := http.ListenAndServe(":"+port, corsMiddleware(router)); err != nil {
```

**何が起きている？**
1. Goサーバーがポート8080でHTTPリクエストを待機
2. リクエストが来ると`router`（Gorilla Mux）が適切なハンドラーを探す

### Step 3: ルーターがURLを解析してハンドラーを決定

**📂 ファイル**: `internal/handler/book_handler.go`

```go
// 373行目 - ルート登録部分
func (h *BookHandler) RegisterRoutes(router *mux.Router) {
    router.HandleFunc("/books", h.CreateBook).Methods("POST")  // ← ここにマッチ！
```

**何が起きている？**
1. URL `/api/v1/books` と HTTPメソッド `POST` の組み合わせを確認
2. `h.CreateBook` 関数が呼び出されることが決定

### Step 4: Handlerでリクエストを処理

**📂 ファイル**: `internal/handler/book_handler.go`

```go
// 52行目～ - CreateBook関数
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    // 56行目：JSONデータを解析
    var req model.CreateBookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.sendErrorResponse(w, http.StatusBadRequest, "リクエストの解析に失敗しました", err)
        return
    }

    // 65行目：UseCaseに処理を依頼
    book, err := h.bookUsecase.CreateBook(&req)
    if err != nil {
        h.sendErrorResponse(w, http.StatusBadRequest, "書籍の作成に失敗しました", err)
        return
    }

    // 73行目：成功レスポンスを返す
    h.sendSuccessResponse(w, http.StatusCreated, "書籍が正常に作成されました", book)
}
```

**何が起きている？**
1. **JSONデコード**: リクエストボディのJSONを`CreateBookRequest`構造体に変換
2. **UseCase呼び出し**: ビジネスロジック層に処理を委譲
3. **レスポンス生成**: 結果をJSON形式でクライアントに返す

### Step 5: UseCaseでビジネスロジックを実行

**📂 ファイル**: `internal/usecase/book_usecase.go`

```go
// 58行目～ - CreateBook関数
func (u *bookUsecase) CreateBook(req *model.CreateBookRequest) (*model.Book, error) {
    // 62行目：入力データのバリデーション
    if err := u.validator.Struct(req); err != nil {
        return nil, fmt.Errorf("入力データが無効です: %w", err)
    }

    // 67行目：ビジネスルール（購入日チェック）
    if req.PurchaseDate.After(time.Now()) {
        return nil, fmt.Errorf("購入日は現在以前の日付を指定してください")
    }

    // 73行目：Repositoryに保存を依頼
    return u.bookRepo.Create(req)
}
```

**何が起きている？**
1. **バリデーション**: 必須項目やデータ形式をチェック
2. **ビジネスルール**: アプリ固有のルール（購入日の未来チェック）を実行
3. **Repository呼び出し**: データアクセス層に処理を委譲

### Step 6: RepositoryでデータベースにSQL実行

**📂 ファイル**: `internal/repository/book_repository.go`

```go
// 42行目～ - Create関数
func (r *bookRepository) Create(req *model.CreateBookRequest) (*model.Book, error) {
    // 47行目：SQL文の定義
    query := `
        INSERT INTO books (title, author, isbn, publisher, published_date, purchase_date, purchase_price, tags, notes)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

    // 55行目：SQLの実行
    result, err := r.db.Exec(query,
        req.Title,         // "Go言語入門"
        req.Author,        // "山田太郎"
        req.ISBN,          // ""
        req.Publisher,     // ""
        req.PublishedDate, // nil
        req.PurchaseDate,  // "2024-01-15T00:00:00Z"
        req.PurchasePrice, // 3000
        req.Tags,          // ""
        req.Notes,         // ""
    )

    // 73行目：挿入されたIDを取得
    id, err := result.LastInsertId()
    
    // 80行目：作成されたデータを取得して返す
    return r.GetByID(int(id))
}
```

**何が起きている？**
1. **SQL生成**: INSERT文でデータベースに新しいレコードを挿入
2. **SQL実行**: SQLiteデータベースに実際にデータを保存
3. **ID取得**: 自動生成されたIDを取得
4. **データ取得**: 保存されたデータを取得して返す

### Step 7: データベースでの実際の処理

**データベース**: `books.db` (SQLite)

```sql
-- 実際に実行されるSQL
INSERT INTO books (title, author, isbn, publisher, published_date, purchase_date, purchase_price, tags, notes)
VALUES ('Go言語入門', '山田太郎', '', '', NULL, '2024-01-15T00:00:00Z', 3000, '', '');

-- 結果: 新しいレコードがID=1で作成される
```

### Step 8: 結果がユーザーに返される

**Handler → UseCase → Repository**の逆順で結果が戻る：

```go
// Repository → UseCase
return &Book{ID: 1, Title: "Go言語入門", ...}, nil

// UseCase → Handler  
return book, nil

// Handler → HTTP Response
{
  "message": "書籍が正常に作成されました",
  "data": {
    "id": 1,
    "title": "Go言語入門",
    "author": "山田太郎",
    "status": "not_started",
    "created_at": "2024-01-15T12:00:00Z"
  }
}
```

## 🔍 実例2: 「書籍一覧を表示する」操作を追跡

### ユーザー操作
```bash
curl "http://localhost:8080/api/v1/books?page=1&limit=10&status=reading"
```

### コードの流れ

1. **ルーター**: `GET /books` → `h.ListBooks`
2. **Handler**: `book_handler.go:106行目`
   ```go
   func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
       // クエリパラメータを解析
       query := r.URL.Query()
       page, _ := strconv.Atoi(query.Get("page"))
       // ...フィルター条件を構築
       books, total, err := h.bookUsecase.ListBooks(filter, page, limit)
   ```

3. **UseCase**: `book_usecase.go:89行目`
   ```go
   func (u *bookUsecase) ListBooks(filter *model.BookFilter, page, limit int) {
       // ページング計算
       offset := (page - 1) * limit
       // Repository呼び出し
       books, err := u.bookRepo.List(filter, limit, offset)
   ```

4. **Repository**: `book_repository.go:136行目`
   ```go
   func (r *bookRepository) List(filter *model.BookFilter, limit, offset int) {
       // 動的SQL構築
       query := "SELECT ... FROM books"
       if filter.Status != nil {
           conditions = append(conditions, "status = ?")
       }
       // SQL実行
       rows, err := r.db.Query(query, args...)
   ```

## 🛠️ デバッグ時の追跡方法

### 1. ログを追加してコードの流れを確認

```go
// Handler層でのログ
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    fmt.Println("=== Handler: CreateBook開始 ===")
    
    var req model.CreateBookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        fmt.Printf("Handler: JSONデコードエラー: %v\n", err)
        return
    }
    fmt.Printf("Handler: リクエストデータ: %+v\n", req)
    
    book, err := h.bookUsecase.CreateBook(&req)
    fmt.Printf("Handler: UseCase結果: book=%+v, err=%v\n", book, err)
}
```

### 2. HTTPリクエスト/レスポンスを確認

```bash
# より詳細な情報を表示
curl -v -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{"title":"テスト"}'
```

### 3. データベースの状態を確認

```bash
# SQLiteを直接確認
sqlite3 books.db "SELECT * FROM books ORDER BY created_at DESC LIMIT 5;"
```

## 🎯 実習：コードの流れを追跡してみよう

### 課題1: 「読書開始」操作の完全追跡

以下のAPIの流れを完全に追跡してください：

```bash
curl -X POST http://localhost:8080/api/v1/books/1/start-reading
```

**追跡すべきポイント**：
1. どのHandlerメソッドが呼ばれる？
2. URLの`{id}`パラメータはどこで取得される？
3. どのUseCaseメソッドが呼ばれる？
4. データベースのどのフィールドが更新される？
5. どんなレスポンスが返される？

### 課題2: エラー時の流れを追跡

```bash
# 存在しないIDで読書開始
curl -X POST http://localhost:8080/api/v1/books/999/start-reading
```

**追跡すべきポイント**：
1. どの層でエラーが発生する？
2. エラーはどのように上位層に伝播する？
3. 最終的にどんなHTTPステータスコードが返される？

### 課題3: フィルタリングの流れを追跡

```bash
curl "http://localhost:8080/api/v1/books?status=completed&rating=5"
```

**追跡すべきポイント**：
1. クエリパラメータはどこで解析される？
2. フィルター条件はどのように構築される？
3. 動的SQLはどのように組み立てられる？

## 📚 次のステップ

1. **実際にログを追加**して、各層での処理を可視化
2. **ブレークポイント**を使ったデバッグ（IDEを使用）
3. **パフォーマンス測定**：各層での処理時間を計測
4. **フロントエンド追加**：実際のWebページからAPIを呼び出す

このドキュメントで、「ボタンを押したら何が起きるのか？」が明確に分かるようになったはずです！🚀
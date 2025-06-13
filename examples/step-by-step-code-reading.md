# 📖 ステップ・バイ・ステップ コードリーディング

このドキュメントでは、**実際のユーザー操作を起点**として、コードを一行ずつ順番に読んでいく方法を説明します。

## 🎯 このドキュメントの目的

- **「このボタンを押すと何が起きるのか？」**を完全に理解する
- **コードを読む順番**を明確にする
- **各行の意味**を詳しく解説する
- **デバッグ時の追跡方法**を身につける

## 📚 実習1: 「書籍を新規作成」の完全解読

### 🚀 実際の操作から開始

```bash
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Go言語プログラミング",
    "author": "山田太郎",
    "purchase_date": "2024-01-15T00:00:00Z",
    "purchase_price": 3500
  }'
```

この操作がどのように処理されるか、**コードを一行ずつ**追っていきましょう。

---

### Step 1: サーバーがリクエストを受信する場所

**📂 ファイル**: `cmd/main.go`

```go
77: log.Printf("サーバーを開始します (ポート: %s)", port)
78: if err := http.ListenAndServe(":"+port, corsMiddleware(router)); err != nil {
79:     log.Fatal("サーバーの開始に失敗しました:", err)
80: }
```

**📝 解説**:
- **77行目**: ログ出力（サーバー開始メッセージ）
- **78行目**: `http.ListenAndServe()` でHTTPサーバーを開始
  - `":"+port` → `:8080` （全てのIPアドレスの8080ポートで待機）
  - `corsMiddleware(router)` → CORSミドルウェアを適用したルーター
- **79行目**: サーバー開始に失敗した場合はプログラム終了

**🔍 ここで何が起きる？**
HTTPリクエストが`corsMiddleware(router)`に渡される

---

### Step 2: CORS処理とルーティング

**📂 ファイル**: `cmd/main.go`

```go
45: func corsMiddleware(handler http.Handler) http.Handler {
46:     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
47:         w.Header().Set("Access-Control-Allow-Origin", "*")
48:         w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
49:         w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
50:         
51:         if r.Method == "OPTIONS" {
52:             return
53:         }
54:         
55:         handler.ServeHTTP(w, r)
56:     })
57: }
```

**📝 解説**:
- **47-49行目**: CORS（Cross-Origin Resource Sharing）ヘッダーを設定
- **51-53行目**: OPTIONSリクエスト（プリフライト）の場合は処理終了
- **55行目**: 実際のハンドラー（router）に処理を移譲

**🔍 ここで何が起きる？**
CORSヘッダーが設定され、`router.ServeHTTP()`が呼ばれる

---

### Step 3: ルーターがURLとメソッドを解析

**📂 ファイル**: `cmd/main.go`

```go
35: // ルートハンドラを設定
36: router := mux.NewRouter().PathPrefix("/api/v1").Subrouter()
37: bookHandler.RegisterRoutes(router)
```

Gorilla Muxが **URL: `/api/v1/books`** と **Method: `POST`** の組み合わせを解析し、適切なハンドラーを探す。

**📂 ファイル**: `internal/handler/book_handler.go`

```go
373: router.HandleFunc("/books", h.CreateBook).Methods("POST")
```

**📝 解説**:
- URL `/api/v1/books` + Method `POST` の組み合わせで `h.CreateBook` 関数が選択される

**🔍 ここで何が起きる？**
`h.CreateBook(w, r)` 関数が呼び出される

---

### Step 4: Handlerでリクエストを解析

**📂 ファイル**: `internal/handler/book_handler.go`

```go
52: func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
53:     // リクエストボディからJSONデータを解析して構造体に変換
54:     var req model.CreateBookRequest
55:     // json.NewDecoder(r.Body).Decode()：HTTPリクエストのJSONをGoの構造体に変換
56:     if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
57:         // パースエラーの場合は400 Bad Requestでエラーレスポンスを返す
58:         h.sendErrorResponse(w, http.StatusBadRequest, "リクエストの解析に失敗しました", err)
59:         return
60:     }
```

**📝 解説**:
- **54行目**: `CreateBookRequest` 構造体の変数を作成（空の状態）
- **56行目**: JSON デコーダーを作成し、HTTPリクエストボディのJSONを構造体に変換
  - `r.Body` → HTTP リクエストボディ（JSON文字列）
  - `&req` → デコード結果を格納する先のアドレス
- **56-60行目**: エラーが発生した場合の処理

**🔍 変換される内容**:
```
JSON: {"title":"Go言語プログラミング","author":"山田太郎",...}
↓
Go構造体: CreateBookRequest{
  Title: "Go言語プログラミング",
  Author: "山田太郎",
  PurchaseDate: time.Time{2024-01-15...},
  PurchasePrice: 3500,
}
```

---

### Step 5: UseCaseに処理を委譲

**📂 ファイル**: `internal/handler/book_handler.go`

```go
62:     // ユースケースでビジネスロジックを実行（バリデーション、データ保存）
63:     book, err := h.bookUsecase.CreateBook(&req)
64:     if err != nil {
65:         // ビジネスロジックエラーの場合は400 Bad Requestでエラーレスポンスを返す
66:         h.sendErrorResponse(w, http.StatusBadRequest, "書籍の作成に失敗しました", err)
67:         return
68:     }
```

**📝 解説**:
- **63行目**: ビジネスロジック層（UseCase）の `CreateBook` メソッドを呼び出し
  - `&req` → 構造体のアドレスを渡す（ポインタ渡し）
  - 戻り値は `(*model.Book, error)` の組み合わせ
- **64-68行目**: エラーが発生した場合の処理

**🔍 ここで何が起きる？**
`h.bookUsecase.CreateBook(&req)` が呼び出される

---

### Step 6: UseCaseでバリデーション実行

**📂 ファイル**: `internal/usecase/book_usecase.go`

```go
58: func (u *bookUsecase) CreateBook(req *model.CreateBookRequest) (*model.Book, error) {
59:     // バリデーション：入力データが正しいかをチェック
60:     // validator.Struct()：構造体のタグ（requiredなど）をチェック
61:     if err := u.validator.Struct(req); err != nil {
62:         return nil, fmt.Errorf("入力データが無効です: %w", err)
63:     }
```

**📝 解説**:
- **61行目**: バリデーターが構造体のタグをチェック
  - `validate:"required"` タグがある項目が空でないかチェック
  - エラーがあれば詳細なエラーメッセージを生成

**🔍 チェックされる内容** (`internal/model/book.go` より):
```go
type CreateBookRequest struct {
    Title         string     `json:"title" validate:"required"`         // ← 必須チェック
    Author        string     `json:"author" validate:"required"`        // ← 必須チェック
    PurchaseDate  time.Time  `json:"purchase_date" validate:"required"` // ← 必須チェック
    // その他のフィールド
}
```

---

### Step 7: ビジネスルールチェック

**📂 ファイル**: `internal/usecase/book_usecase.go`

```go
65:     // ビジネスルール：購入日が未来でないことを確認
66:     // time.Now().After()：指定した時刻より後かどうかを判定
67:     if req.PurchaseDate.After(time.Now()) {
68:         return nil, fmt.Errorf("購入日は現在以前の日付を指定してください")
69:     }
```

**📝 解説**:
- **67行目**: 購入日が現在時刻より後（未来）でないかをチェック
  - `req.PurchaseDate` → `2024-01-15T00:00:00Z`
  - `time.Now()` → 現在時刻
  - `.After()` → 前者が後者より後かを判定

**🔍 ここで何が起きる？**
購入日が未来の場合はエラーを返し、正常な場合は次に進む

---

### Step 8: Repositoryに保存を依頼

**📂 ファイル**: `internal/usecase/book_usecase.go`

```go
71:     // 検証が成功したらリポジトリに作成を依頼
72:     return u.bookRepo.Create(req)
```

**📝 解説**:
- **72行目**: データアクセス層（Repository）の `Create` メソッドを呼び出し
- 戻り値をそのまま返す（正常時: `*model.Book`, エラー時: `error`）

**🔍 ここで何が起きる？**
`u.bookRepo.Create(req)` が呼び出される

---

### Step 9: SQLクエリの準備と実行

**📂 ファイル**: `internal/repository/book_repository.go`

```go
42: func (r *bookRepository) Create(req *model.CreateBookRequest) (*model.Book, error) {
43:     // query：SQL文（データベースに実行させる命令）
44:     // INSERT INTO：新しいデータを挿入するSQL命令
45:     // ?：プレースホルダー（後で実際の値に置き換えられる）
46:     query := `
47:         INSERT INTO books (title, author, isbn, publisher, published_date, purchase_date, purchase_price, tags, notes)
48:         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
49:     `
```

**📝 解説**:
- **46-49行目**: SQL INSERT文を文字列として定義
- `?` はプレースホルダー（SQLインジェクション対策）
- 9個の`?`が9個のカラムに対応

**🔍 実際のSQL**:
```sql
INSERT INTO books (title, author, isbn, publisher, published_date, purchase_date, purchase_price, tags, notes)
VALUES ('Go言語プログラミング', '山田太郎', '', '', NULL, '2024-01-15T00:00:00Z', 3500, '', '')
```

---

### Step 10: SQLの実行

**📂 ファイル**: `internal/repository/book_repository.go`

```go
51:     // r.db.Exec()：SQLを実行する関数
52:     // プレースホルダー（?）に実際の値を順番に入れて実行
53:     result, err := r.db.Exec(query,
54:         req.Title,         // "Go言語プログラミング"
55:         req.Author,        // "山田太郎"  
56:         req.ISBN,          // ""
57:         req.Publisher,     // ""
58:         req.PublishedDate, // nil
59:         req.PurchaseDate,  // time.Time{2024-01-15...}
60:         req.PurchasePrice, // 3500
61:         req.Tags,          // ""
62:         req.Notes,         // ""
63:     )
```

**📝 解説**:
- **53-63行目**: SQL文を実行し、プレースホルダーに実際の値を代入
- `r.db.Exec()` は SQLiteドライバーの関数
- 戻り値: `result` (実行結果), `err` (エラー)

**🔍 データベースでの処理**:
1. SQL文が解析される
2. プレースホルダーに値が安全に代入される
3. books テーブルに新しいレコードが挿入される
4. 自動生成されたID（例：1）がレコードに設定される

---

### Step 11: 挿入されたIDを取得

**📂 ファイル**: `internal/repository/book_repository.go`

```go
64:     // エラーハンドリング：エラーが発生した場合の処理
65:     if err != nil {
66:         return nil, fmt.Errorf("書籍の作成に失敗しました: %w", err)
67:     }
68: 
69:     // LastInsertId()：挿入されたデータの自動生成ID（主キー）を取得
70:     id, err := result.LastInsertId()
71:     if err != nil {
72:         return nil, fmt.Errorf("書籍IDの取得に失敗しました: %w", err)
73:     }
```

**📝 解説**:
- **65-67行目**: SQL実行エラーのチェック
- **70行目**: 挿入されたレコードの自動生成ID（主キー）を取得
  - SQLiteの場合、`id` カラムが自動でインクリメントされる
  - `LastInsertId()` でその値を取得（例：1）

---

### Step 12: 作成されたデータを取得

**📂 ファイル**: `internal/repository/book_repository.go`

```go
75:     // 作成された書籍のデータを取得して返す
76:     // int(id)：int64型をint型に変換
77:     return r.GetByID(int(id))
```

**📝 解説**:
- **77行目**: 挿入されたIDで書籍データを再取得
  - `id` は `int64` 型なので `int` 型に変換
  - `GetByID()` メソッドを呼び出し

**🔍 ここで何が起きる？**
`r.GetByID(1)` が呼び出される（IDが1の場合）

---

### Step 13: データ取得用SQLの実行

**📂 ファイル**: `internal/repository/book_repository.go`

```go
84: func (r *bookRepository) GetByID(id int) (*model.Book, error) {
85:     // SELECT文：booksテーブルから指定したカラム（列）のデータを取得
86:     // WHERE id = ?：IDが一致する行だけを取得する条件
87:     query := `
88:         SELECT id, title, author, isbn, publisher, published_date, purchase_date, 
89:                purchase_price, status, start_read_date, end_read_date, rating, 
90:                notes, tags, created_at, updated_at
91:         FROM books 
92:         WHERE id = ?
93:     `
94: 
95:     // &model.Book{}：空のBook構造体を作成（&でポインタにする）
96:     book := &model.Book{}
97:     // QueryRow()：1行だけを取得するSQL実行関数
98:     row := r.db.QueryRow(query, id)
```

**📝 解説**:
- **87-93行目**: SELECT文でIDに一致するレコードを1件取得
- **96行目**: 空の `Book` 構造体を作成
- **98行目**: `QueryRow()` で1行分のクエリを実行

---

### Step 14: 取得データを構造体に格納

**📂 ファイル**: `internal/repository/book_repository.go`

```go
100:    // Scan()：取得したデータを構造体の各フィールドに格納
101:    // &book.ID：bookのIDフィールドのアドレス（格納先を指定）
102:    err := row.Scan(
103:        &book.ID,            // 1
104:        &book.Title,         // "Go言語プログラミング"
105:        &book.Author,        // "山田太郎"
106:        &book.ISBN,          // ""
107:        &book.Publisher,     // ""
108:        &book.PublishedDate, // nil
109:        &book.PurchaseDate,  // time.Time{2024-01-15...}
110:        &book.PurchasePrice, // 3500
111:        &book.Status,        // "not_started"
112:        &book.StartReadDate, // nil
113:        &book.EndReadDate,   // nil
114:        &book.Rating,        // nil
115:        &book.Notes,         // ""
116:        &book.Tags,          // ""
117:        &book.CreatedAt,     // time.Time{2024-01-15 12:00:00...}
118:        &book.UpdatedAt,     // time.Time{2024-01-15 12:00:00...}
119:    )
```

**📝 解説**:
- **102-119行目**: データベースの各カラムの値を、`Book` 構造体の各フィールドに格納
- `&book.ID` のように `&` をつけることで、「この場所に値を格納してください」と指定

**🔍 結果**:
```go
book := &model.Book{
    ID: 1,
    Title: "Go言語プログラミング",
    Author: "山田太郎",
    Status: "not_started",
    PurchaseDate: time.Time{2024-01-15...},
    PurchasePrice: 3500,
    CreatedAt: time.Time{2024-01-15 12:00:00...},
    // その他のフィールド
}
```

---

### Step 15: 結果を上位層に返す

データが正常に作成・取得できたので、結果を上位層に順番に返していきます：

**Repository → UseCase**:
```go
// repository/book_repository.go:131行目
return book, nil  // 正常に作成された book を返す
```

**UseCase → Handler**:
```go
// usecase/book_usecase.go:72行目  
return u.bookRepo.Create(req)  // Repository の結果をそのまま返す
```

**Handler → HTTP Response**:

**📂 ファイル**: `internal/handler/book_handler.go`

```go
70:     // 成功時は201 Createdで作成された書籍データを返す
71:     h.sendSuccessResponse(w, http.StatusCreated, "書籍が正常に作成されました", book)
```

---

### Step 16: JSONレスポンスの生成

**📂 ファイル**: `internal/handler/book_handler.go`

```go
350: func (h *BookHandler) sendSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
351:     // HTTPレスポンスヘッダーを設定（JSON形式で返すことを明示）
352:     w.Header().Set("Content-Type", "application/json")
353:     // HTTPステータスコードを設定（200, 201など）
354:     w.WriteHeader(statusCode)
355: 
356:     // 成功レスポンス構造体を作成
357:     response := SuccessResponse{
358:         Message: message, // "書籍が正常に作成されました"
359:         Data:    data,    // book データ
360:     }
361: 
362:     // JSON形式でレスポンスを送信
363:     json.NewEncoder(w).Encode(response)
364: }
```

**📝 解説**:
- **352行目**: HTTP ヘッダーに `Content-Type: application/json` を設定
- **354行目**: HTTP ステータスコード `201 Created` を設定
- **357-360行目**: レスポンス用の構造体を作成
- **363行目**: Go構造体をJSON文字列に変換してHTTPレスポンスとして送信

**🔍 最終的なJSONレスポンス**:
```json
{
  "message": "書籍が正常に作成されました",
  "data": {
    "id": 1,
    "title": "Go言語プログラミング",
    "author": "山田太郎",
    "isbn": "",
    "publisher": "",
    "published_date": null,
    "purchase_date": "2024-01-15T00:00:00Z",
    "purchase_price": 3500,
    "status": "not_started",
    "start_read_date": null,
    "end_read_date": null,
    "rating": null,
    "notes": "",
    "tags": "",
    "created_at": "2024-01-15T12:00:00Z",
    "updated_at": "2024-01-15T12:00:00Z"
  }
}
```

---

## 🎯 まとめ：データの流れ

```
1. HTTP Request (JSON)
   ↓
2. Gorilla Mux Router
   ↓  
3. Handler (JSON → Go struct)
   ↓
4. UseCase (Validation + Business Logic)
   ↓
5. Repository (Go struct → SQL)
   ↓
6. SQLite Database (INSERT)
   ↓
7. Database (SELECT)
   ↓
8. Repository (SQL → Go struct)
   ↓
9. UseCase (pass through)
   ↓
10. Handler (Go struct → JSON)
    ↓
11. HTTP Response (JSON)
```

## 🔍 実習課題

### 課題1: 書籍取得APIを追跡
```bash
curl http://localhost:8080/api/v1/books/1
```
上記APIのコードの流れを、この例と同様に一行ずつ追跡してみてください。

### 課題2: エラー時の流れを追跡  
```bash
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{"title":"タイトルのみ"}'  # authorが不足
```
バリデーションエラーが発生する場合の流れを追跡してみてください。

このような方法で、**どんな操作でも一行ずつコードを追跡**できるようになります！🚀
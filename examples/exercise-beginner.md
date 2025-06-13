# 🏋️ 初心者向け実習課題

このファイルには、書籍管理アプリを使った実習課題を載せています。**段階的に難易度が上がる**ように設計されているので、順番に挑戦してみてください！

## 🎯 実習の進め方

1. **まずはアプリを動かす**：`go run cmd/main.go`
2. **課題を読む**
3. **実際にコードを変更する**
4. **動作確認する**
5. **次の課題へ**

## 🚀 Level 1: アプリの基本操作

### 課題1-1: アプリを起動してみよう
```bash
# アプリを起動
go run cmd/main.go

# 別のターミナルで動作確認
curl http://localhost:8080/api/v1/health
```

**期待する結果**：`{"message":"サービスは正常に動作しています",...}`

### 課題1-2: 初めての書籍を登録してみよう
```bash
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "あなたの好きな本のタイトル",
    "author": "著者名",
    "purchase_date": "2024-01-15T00:00:00Z",
    "purchase_price": 3000
  }'
```

### 課題1-3: 登録した書籍を確認してみよう
```bash
# 書籍一覧を確認
curl http://localhost:8080/api/v1/books

# 特定の書籍を確認（IDは1と仮定）
curl http://localhost:8080/api/v1/books/1
```

### 課題1-4: 読書を開始して完了してみよう
```bash
# 読書開始
curl -X POST http://localhost:8080/api/v1/books/1/start-reading

# 読書完了（5点評価）
curl -X POST http://localhost:8080/api/v1/books/1/finish-reading \
  -H "Content-Type: application/json" \
  -d '{"rating": 5}'

# 統計情報を確認
curl http://localhost:8080/api/v1/statistics
```

## 🛠️ Level 2: コードを読んで理解しよう

### 課題2-1: main.goを読んでみよう
`cmd/main.go`を開いて、以下の質問に答えてください：

1. アプリは何番ポートで起動する？
2. データベースファイルはどこに作られる？
3. `defer`文は何をしている？

### 課題2-2: Book構造体を理解しよう
`internal/model/book.go`を開いて、以下の質問に答えてください：

1. 書籍に設定できる読書ステータスは何種類？
2. 評価（Rating）フィールドが`*int`（ポインタ）になっているのはなぜ？
3. `json:"title"`のタグは何のため？

### 課題2-3: API処理の流れを追ってみよう
書籍作成（POST /api/v1/books）の処理が以下のファイルのどの関数で行われているか確認してください：

1. `internal/handler/book_handler.go` → ?
2. `internal/usecase/book_usecase.go` → ?
3. `internal/repository/book_repository.go` → ?

## ✏️ Level 3: 簡単なカスタマイズ

### 課題3-1: 成功メッセージを変更してみよう
`internal/handler/book_handler.go`の`CreateBook`関数で、成功時のメッセージを変更してみてください。

**変更前**：
```go
h.sendSuccessResponse(w, http.StatusCreated, "書籍が正常に作成されました", book)
```

**変更後**：
```go
h.sendSuccessResponse(w, http.StatusCreated, "新しい本が追加されました！📚", book)
```

**確認方法**：
```bash
# アプリを再起動して、書籍作成APIを実行
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{"title":"テスト","author":"テスト","purchase_date":"2024-01-15T00:00:00Z"}'
```

### 課題3-2: ヘルスチェックのレスポンスを変更してみよう
`internal/handler/book_handler.go`の`Health`関数を変更して、あなたの名前を含むメッセージに変更してください。

### 課題3-3: 新しい読書ステータスを追加してみよう
「ほしい物リスト」機能として、`wish_list`ステータスを追加してみてください。

**変更するファイル**：
1. `internal/model/book.go` - 新しいステータス定数を追加
2. `internal/database/migration.sql` - SQLiteのCHECK制約を更新

## 🏗️ Level 4: 新機能を追加

### 課題4-1: 書籍にページ数フィールドを追加
書籍にページ数（`page_count`）フィールドを追加してみましょう。

**手順**：
1. `internal/model/book.go`の`Book`構造体にフィールド追加
2. `CreateBookRequest`と`UpdateBookRequest`にも追加
3. `internal/database/migration.sql`にカラム追加
4. `internal/repository/book_repository.go`のSQL文を更新

**追加するフィールド**：
```go
PageCount int `json:"page_count" db:"page_count"`
```

**SQLでの追加**：
```sql
ALTER TABLE books ADD COLUMN page_count INTEGER DEFAULT 0;
```

### 課題4-2: 「お気に入り」機能を追加
書籍を「お気に入り」に登録する機能を追加してみましょう。

**必要な変更**：
1. `Book`構造体に`IsFavorite bool`フィールド追加
2. データベースにカラム追加
3. お気に入り切り替えAPIの追加

**新しいAPIエンドポイント**：
```
POST /api/v1/books/{id}/toggle-favorite
```

### 課題4-3: 読書時間記録機能
読書にかかった時間を記録する機能を追加してみましょう。

**必要な機能**：
1. 読書時間（分）を記録するフィールド
2. 開始日時と終了日時から自動計算
3. 統計情報に「総読書時間」を追加

## 🧠 Level 5: アーキテクチャ理解

### 課題5-1: 新しいAPIエンドポイントを作成
「書籍を中断する」APIを作成してください。

**要件**：
- エンドポイント：`POST /api/v1/books/{id}/drop-reading`
- ステータスを`dropped`に変更
- 終了日時を記録

**作成する関数**：
1. `BookHandler.DropReading`
2. `BookUsecase.DropReading`
3. ルーターへの登録

### 課題5-2: バリデーション機能を強化
以下のバリデーションを追加してください：

1. **タイトルの長さ制限**：1文字以上100文字以下
2. **価格の範囲**：0円以上100万円以下
3. **ISBNの形式チェック**：13桁の数字

### 課題5-3: データベース関数の追加
「著者別の統計」を取得する機能を追加してください。

**必要な機能**：
1. 著者ごとの書籍数
2. 著者ごとの平均評価
3. 著者ごとの総支出額

## 🚀 Level 6: 応用課題

### 課題6-1: ログ機能の追加
APIアクセスのログを記録する機能を追加してください。

**要件**：
- アクセス時間、IPアドレス、HTTPメソッド、URLを記録
- ログファイルに保存

### 課題6-2: 設定ファイル対応
環境変数ではなく、設定ファイル（JSON/YAML）から設定を読み込む機能を追加してください。

### 課題6-3: テストコードの作成
以下のテストコードを作成してください：

1. **Unit Test**：`BookUsecase.CreateBook`のテスト
2. **Integration Test**：API全体のテスト
3. **Test Data**：テスト用のデータベース作成

## 💡 ヒントとコツ

### デバッグのコツ
```go
// デバッグ出力を追加
fmt.Printf("Debug: book = %+v\n", book)

// エラーの詳細を確認
if err != nil {
    fmt.Printf("Error details: %+v\n", err)
    return nil, err
}
```

### データベースの確認方法
```bash
# SQLiteファイルを直接確認
sqlite3 books.db ".schema books"
sqlite3 books.db "SELECT * FROM books;"
```

### JSON形式の確認
```bash
# 整形して表示
curl http://localhost:8080/api/v1/books | jq .
```

## 🎯 評価基準

### レベル1-2: 基本操作 ⭐
- アプリの起動と基本API操作ができる
- コードを読んで基本的な構造を理解できる

### レベル3: カスタマイズ ⭐⭐
- 既存コードを安全に変更できる
- 変更の影響範囲を理解できる

### レベル4: 機能追加 ⭐⭐⭐
- 新しいフィールドやAPIを追加できる
- データベースの変更も含めて実装できる

### レベル5: アーキテクチャ ⭐⭐⭐⭐
- クリーンアーキテクチャを理解している
- 適切な層に機能を実装できる

### レベル6: 応用 ⭐⭐⭐⭐⭐
- 実際のWebアプリケーションに必要な機能を追加できる
- テストコードも含めて高品質な実装ができる

## 🆘 困った時は

1. **エラーメッセージをよく読む**：どこで何が起きているかを確認
2. **ログを確認**：コンソール出力やエラーログをチェック
3. **小さく分けて試す**：一度に大きな変更をせず、少しずつ確認
4. **元に戻す**：うまくいかない場合は`git checkout`で元に戻す

頑張って挑戦してみてください！🚀

---

**📚 参考資料**
- [Go公式ドキュメント](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [JSON and Go](https://blog.golang.org/json-and-go)
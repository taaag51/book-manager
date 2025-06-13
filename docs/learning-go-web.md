# 📚 Go言語とWeb開発を学ぼう！

このドキュメントは、**プログラミング超初心者**のためのGo言語とWeb開発学習ガイドです。

## 🎯 このプロジェクトで学べること

### 1. Go言語の基本
- **パッケージ（package）**：機能をまとめるフォルダのようなもの
- **関数（func）**：処理をまとめたブロック
- **構造体（struct）**：複数のデータをセットにしたもの
- **ポインタ（*）**：データの場所を示す住所のようなもの
- **インターフェース（interface）**：「こんな機能を持つ」という約束

### 2. Webアプリケーションの仕組み
- **HTTP**：ブラウザとサーバーが話すための言語
- **REST API**：アプリ同士がデータをやり取りする方法
- **JSON**：データを送受信する時の形式
- **データベース**：データを永続的に保存する場所

### 3. クリーンアーキテクチャ
このプロジェクトでは、以下の4つの層に分けてコードを整理しています：

```
ブラウザ → Handler → UseCase → Repository → データベース
```

- **Handler**：ブラウザからのリクエストを受け取る担当
- **UseCase**：ビジネスルール（「評価は1-5点まで」など）を処理する担当
- **Repository**：データベースとやり取りする担当
- **Model**：データの形を決める担当

## 🚀 ステップバイステップ学習

### Step 1: main.goを理解しよう
まずは `cmd/main.go` を見てみましょう。

```go
func main() {
    // アプリケーションの開始点
    // データベースに接続 → ルートを設定 → サーバー開始
}
```

**学習ポイント**：
- `main`関数はアプリの開始点
- `defer`は関数終了時に実行される処理
- `log.Fatal`はエラーが起きたらアプリを停止

### Step 2: データの形を理解しよう
`internal/model/book.go` を見てみましょう。

```go
type Book struct {
    ID     int    `json:"id"`     // 書籍のID番号
    Title  string `json:"title"`  // タイトル
    Author string `json:"author"` // 著者
}
```

**学習ポイント**：
- `struct`は複数のデータをまとめる箱
- `json:"xxx"`はAPI通信時の項目名
- `*int`の`*`は「値がない可能性がある」という意味

### Step 3: データベース操作を理解しよう
`internal/repository/book_repository.go` を見てみましょう。

```go
func (r *bookRepository) Create(req *model.CreateBookRequest) (*model.Book, error) {
    // 1. SQL文を作る
    // 2. データベースに実行させる
    // 3. 結果を返す
}
```

**学習ポイント**：
- SQL文でデータベースに命令を出す
- `?`は後で値を入れる場所（プレースホルダー）
- エラーハンドリングは毎回必要

### Step 4: ビジネスロジックを理解しよう
`internal/usecase/book_usecase.go` を見てみましょう。

```go
func (u *bookUsecase) CreateBook(req *model.CreateBookRequest) (*model.Book, error) {
    // 1. 入力データをチェック
    // 2. ビジネスルールをチェック（購入日が未来でないかなど）
    // 3. データベースに保存
}
```

**学習ポイント**：
- バリデーション（検証）は重要
- ビジネスルールはここで実装
- 複雑な計算や統計処理もここ

### Step 5: HTTP処理を理解しよう
`internal/handler/book_handler.go` を見てみましょう。

```go
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    // 1. リクエストのJSONを解析
    // 2. ユースケースに処理を依頼
    // 3. 結果をJSONで返す
}
```

**学習ポイント**：
- HTTPハンドラーはリクエストとレスポンスを処理
- JSONエンコード/デコードでデータ変換
- エラー時は適切なステータスコードを返す

## 🔧 実際に試してみよう

### 1. APIを叩いてみる

アプリを起動後、以下のコマンドを試してみてください：

```bash
# 1. 健康状態を確認
curl http://localhost:8080/api/v1/health

# 2. 書籍を1冊登録
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "はじめてのGo言語",
    "author": "山田太郎",
    "purchase_date": "2024-01-15T00:00:00Z",
    "purchase_price": 3000
  }'

# 3. 書籍一覧を確認
curl http://localhost:8080/api/v1/books

# 4. 読書を開始（IDは1番と仮定）
curl -X POST http://localhost:8080/api/v1/books/1/start-reading

# 5. 読書を完了
curl -X POST http://localhost:8080/api/v1/books/1/finish-reading \
  -H "Content-Type: application/json" \
  -d '{"rating": 5}'

# 6. 統計情報を確認
curl http://localhost:8080/api/v1/statistics
```

### 2. コードを変更してみる

以下の簡単な変更を試してみてください：

#### 🎯 初級：メッセージを変更
`internal/handler/book_handler.go` の成功メッセージを変更してみましょう。

```go
// 変更前
h.sendSuccessResponse(w, http.StatusCreated, "書籍が正常に作成されました", book)

// 変更後
h.sendSuccessResponse(w, http.StatusCreated, "新しい本を追加しました！", book)
```

#### 🎯 中級：新しいフィールドを追加
書籍にページ数フィールドを追加してみましょう。

1. `internal/model/book.go` にフィールド追加
2. `internal/database/migration.sql` にカラム追加
3. Repository、UseCase、Handlerを更新

#### 🎯 上級：新しいAPIエンドポイントを追加
「読書を中断する」機能を追加してみましょう。

## 📖 さらに学習を深めるために

### 推奨学習順序
1. **Go言語の基本文法**：[公式チュートリアル](https://tour.golang.org/)
2. **HTTP/Web基礎**：HTTPメソッド、ステータスコード、JSONについて
3. **データベース基礎**：SQL文の基本（SELECT、INSERT、UPDATE、DELETE）
4. **アーキテクチャパターン**：なぜコードを層に分けるのか

### 参考資料
- [Go言語公式ドキュメント](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [JSON and Go](https://blog.golang.org/json-and-go)

## 🤔 よくある質問

### Q: ポインタ（*）って何ですか？
A: データの住所を表すものです。「実際の値」ではなく「値がある場所」を指します。

```go
var x int = 42
var p *int = &x  // pはxの住所を持つ
fmt.Println(*p)  // 42 (*pでその住所にある値を取得)
```

### Q: なぜエラーハンドリングが毎回必要なんですか？
A: Goでは「エラーが起きる可能性がある処理」は必ずエラーをチェックします。これにより、問題を早期発見できます。

```go
data, err := someFunction()
if err != nil {
    // エラーが起きた時の処理
    return nil, err
}
// エラーがない時の処理
```

### Q: SQLインジェクションって何ですか？
A: 悪意あるSQL文を注入される攻撃です。プレースホルダー（?）を使うことで防げます。

```go
// 危険（SQLインジェクション脆弱性あり）
query := "SELECT * FROM books WHERE title = '" + title + "'"

// 安全（プレースホルダーを使用）
query := "SELECT * FROM books WHERE title = ?"
```

頑張って学習を続けてください！🚀
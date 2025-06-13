# 🐛 デバッグ・トラブルシューティングガイド

このドキュメントでは、書籍管理アプリの開発中に起こりがちな問題と、その解決方法を説明します。

## 🎯 このガイドの使い方

1. **問題が起きたら症状を確認**
2. **該当する問題パターンを探す**
3. **解決手順を順番に実行**
4. **予防策を覚える**

## 🚨 よくある問題と解決方法

### 1. アプリが起動しない

#### 症状
```bash
$ go run cmd/main.go
# 何も表示されない、またはエラーメッセージが出る
```

#### パターン1: ポートが既に使用されている
```
listen tcp :8080: bind: address already in use
```

**解決方法**：
```bash
# 1. ポート8080を使っているプロセスを確認
lsof -i :8080

# 2. プロセスを終了
kill -9 [プロセスID]

# 3. または別のポートを使用
PORT=8081 go run cmd/main.go
```

#### パターン2: データベースファイルの権限エラー
```
failed to open database: permission denied
```

**解決方法**：
```bash
# 1. 現在のディレクトリの権限を確認
ls -la

# 2. 書き込み権限を追加
chmod 755 .

# 3. または明示的にパスを指定
DB_PATH=/tmp/books.db go run cmd/main.go
```

#### パターン3: 依存関係の問題
```
cannot find module for path
```

**解決方法**：
```bash
# 1. モジュールの整理
go mod tidy

# 2. 依存関係の再取得
go mod download

# 3. モジュールキャッシュのクリア
go clean -modcache
```

### 2. APIが期待通りに動作しない

#### 症状
```bash
$ curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{"title":"テスト"}'

# エラーレスポンスまたは期待と違う結果
```

#### パターン1: 必須フィールドが不足
```json
{
  "error": "入力データが無効です",
  "message": "Author is required"
}
```

**解決方法**：
```bash
# model/book.goでrequiredフィールドを確認
grep -n "validate:\"required\"" internal/model/book.go

# 正しいリクエストを送信
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "テスト書籍",
    "author": "テスト著者",
    "purchase_date": "2024-01-15T00:00:00Z"
  }'
```

#### パターン2: 日付形式のエラー
```json
{
  "error": "リクエストの解析に失敗しました",
  "message": "parsing time \"2024-01-15\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"\" as \"T\""
}
```

**解決方法**：
```bash
# RFC3339形式（ISO8601）で日付を指定
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "テスト",
    "author": "テスト",
    "purchase_date": "2024-01-15T00:00:00Z"
  }'
```

#### パターン3: Content-Typeヘッダーの不足
```json
{
  "error": "リクエストの解析に失敗しました"
}
```

**解決方法**：
```bash
# Content-Typeヘッダーを必ず追加
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{"title":"テスト","author":"テスト","purchase_date":"2024-01-15T00:00:00Z"}'
```

### 3. データベース関連の問題

#### 症状
```
database is locked
sql: no rows in result set
```

#### パターン1: データベースファイルが破損
**解決方法**：
```bash
# 1. アプリを停止
# 2. データベースファイルを削除
rm books.db

# 3. アプリを再起動（自動でテーブルが作成される）
go run cmd/main.go
```

#### パターン2: SQLクエリのエラー
**デバッグ方法**：
```bash
# 1. SQLiteで直接確認
sqlite3 books.db

# 2. テーブル構造を確認
.schema books

# 3. データを確認
SELECT * FROM books LIMIT 5;

# 4. 終了
.quit
```

### 4. コード変更後に問題が発生

#### 症状
```
undefined: SomeFunction
cannot use X as Y in assignment
```

#### パターン1: インポートの不足
**解決方法**：
```bash
# 1. 自動でインポートを整理
go mod tidy

# 2. goimportsツールを使用（推奨）
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

#### パターン2: 型の不一致
**デバッグ方法**：
```go
// デバッグ用の型確認
fmt.Printf("変数の型: %T, 値: %+v\n", variable, variable)

// ポインタのnilチェック
if ptr == nil {
    fmt.Println("ポインタがnilです")
}
```

## 🔍 効果的なデバッグ手法

### 1. ログを活用したデバッグ

```go
// 各層でのログ追加例

// Handler層
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    log.Printf("[Handler] CreateBook開始")
    
    var req model.CreateBookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("[Handler] JSONデコードエラー: %v", err)
        h.sendErrorResponse(w, http.StatusBadRequest, "リクエストの解析に失敗しました", err)
        return
    }
    log.Printf("[Handler] リクエストデータ: %+v", req)
    
    book, err := h.bookUsecase.CreateBook(&req)
    if err != nil {
        log.Printf("[Handler] UseCase エラー: %v", err)
        h.sendErrorResponse(w, http.StatusBadRequest, "書籍の作成に失敗しました", err)
        return
    }
    log.Printf("[Handler] 作成成功: %+v", book)
    
    h.sendSuccessResponse(w, http.StatusCreated, "書籍が正常に作成されました", book)
}

// UseCase層
func (u *bookUsecase) CreateBook(req *model.CreateBookRequest) (*model.Book, error) {
    log.Printf("[UseCase] CreateBook開始: %+v", req)
    
    if err := u.validator.Struct(req); err != nil {
        log.Printf("[UseCase] バリデーションエラー: %v", err)
        return nil, fmt.Errorf("入力データが無効です: %w", err)
    }
    
    book, err := u.bookRepo.Create(req)
    if err != nil {
        log.Printf("[UseCase] Repository エラー: %v", err)
        return nil, err
    }
    log.Printf("[UseCase] 作成成功: %+v", book)
    
    return book, nil
}
```

### 2. HTTPリクエスト/レスポンスの詳細確認

```bash
# より詳細な情報を表示
curl -v -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{"title":"テスト","author":"テスト","purchase_date":"2024-01-15T00:00:00Z"}'

# レスポンスを整形して表示
curl -X GET http://localhost:8080/api/v1/books | jq .

# レスポンスヘッダーも確認
curl -I http://localhost:8080/api/v1/health
```

### 3. 段階的なテスト

```bash
# 1. まずヘルスチェック
curl http://localhost:8080/api/v1/health

# 2. 簡単なGETリクエスト
curl http://localhost:8080/api/v1/books

# 3. 最小限のPOSTリクエスト
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{"title":"最小テスト","author":"テスト","purchase_date":"2024-01-15T00:00:00Z"}'

# 4. 複雑なリクエスト
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "完全版テスト",
    "author": "テスト著者",
    "isbn": "978-4-123-45678-9",
    "publisher": "テスト出版",
    "published_date": "2023-06-15T00:00:00Z",
    "purchase_date": "2024-01-15T00:00:00Z",
    "purchase_price": 3000,
    "tags": "プログラミング,テスト",
    "notes": "テスト用の書籍です"
  }'
```

## 🛠️ 開発環境のセットアップ確認

### Go環境のチェック
```bash
# Goバージョン確認
go version  # 1.21以上が必要

# 作業ディレクトリ確認
pwd
ls -la

# モジュール確認
go mod verify
```

### エディタ/IDE設定確認
```bash
# VS Codeの場合：Go拡張機能の確認
code --list-extensions | grep golang

# 自動フォーマット確認
go fmt ./...

# 静的解析実行
go vet ./...
```

## 🧪 テスト用ツールとコマンド

### データベース確認ツール
```bash
# SQLiteクライアント
sqlite3 books.db

# よく使うSQLコマンド
.tables                          # テーブル一覧
.schema books                    # テーブル構造
SELECT COUNT(*) FROM books;      # レコード数
SELECT * FROM books ORDER BY created_at DESC LIMIT 10;  # 最新10件
```

### API テストツール
```bash
# curl以外の選択肢

# httpie（より人間にやさしい）
http POST localhost:8080/api/v1/books title="テスト" author="テスト" purchase_date="2024-01-15T00:00:00Z"

# jq（JSONの整形・フィルタリング）
curl -s http://localhost:8080/api/v1/books | jq '.data.books[0].title'
```

## 📋 トラブルシューティング チェックリスト

### 問題が起きたら以下を順番に確認

#### ✅ 基本確認
- [ ] アプリは起動している？（`curl http://localhost:8080/api/v1/health`）
- [ ] ポート番号は正しい？（デフォルト：8080）
- [ ] HTTPメソッドは正しい？（GET/POST/PUT/DELETE）
- [ ] URLパスは正しい？（`/api/v1/books`）

#### ✅ リクエスト確認
- [ ] Content-Typeヘッダーは設定されている？
- [ ] JSONの形式は正しい？（波括弧、引用符、カンマ）
- [ ] 必須フィールドは全て含まれている？
- [ ] 日付の形式はRFC3339？（`2024-01-15T00:00:00Z`）

#### ✅ レスポンス確認
- [ ] ステータスコードは何？（200/201/400/404/500）
- [ ] エラーメッセージの内容は？
- [ ] レスポンスボディにデータは含まれている？

#### ✅ コード確認
- [ ] 変更したファイルは保存されている？
- [ ] アプリの再起動は済んでいる？
- [ ] importは正しく設定されている？
- [ ] 型の不一致はない？

#### ✅ データベース確認
- [ ] `books.db`ファイルは存在する？
- [ ] テーブルは正しく作成されている？
- [ ] データは正しく挿入されている？

## 🆘 それでも解決しない場合

### 1. ログを詳細化
```go
// より詳細なログを追加
log.SetFlags(log.LstdFlags | log.Lshortfile)  // ファイル名と行番号を表示
```

### 2. 最小構成で再現
```go
// 最小限のテストケースを作成
func main() {
    // 問題のある部分だけを切り出してテスト
}
```

### 3. 元の状態に戻す
```bash
# Gitで最後の正常な状態に戻す
git checkout HEAD -- .
```

### 4. 段階的に機能を戻す
- 一度に大きく変更せず、小さな変更を積み重ねる
- 動作確認しながら一歩ずつ進める

このガイドを使って、効率的にデバッグを進めてください！🚀
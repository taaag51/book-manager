# 📡 API使用例集

このファイルには、書籍管理アプリのAPIを実際に使う例を載せています。

## 🎯 基本的な使い方の流れ

### 1. アプリケーションを起動
```bash
cd /path/to/ccode-sample
go run cmd/main.go
```

### 2. ブラウザでアクセス
- メインページ: http://localhost:8080
- ヘルスチェック: http://localhost:8080/api/v1/health

## 📚 書籍管理の基本操作

### ✅ 1. 書籍を1冊追加する

```bash
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Go言語でつくるWebアプリケーション",
    "author": "田中太郎",
    "isbn": "978-4-123-45678-9",
    "publisher": "技術評論社",
    "published_date": "2023-06-15T00:00:00Z",
    "purchase_date": "2023-12-01T00:00:00Z",
    "purchase_price": 3200,
    "tags": "プログラミング,Go言語,Web開発",
    "notes": "初心者にも分かりやすい良書"
  }'
```

**期待される結果**：
```json
{
  "message": "書籍が正常に作成されました",
  "data": {
    "id": 1,
    "title": "Go言語でつくるWebアプリケーション",
    "author": "田中太郎",
    "status": "not_started",
    "created_at": "2024-01-15T12:00:00Z"
  }
}
```

### 📖 2. すべての書籍を見る

```bash
curl http://localhost:8080/api/v1/books
```

**期待される結果**：
```json
{
  "message": "",
  "data": {
    "books": [
      {
        "id": 1,
        "title": "Go言語でつくるWebアプリケーション",
        "status": "not_started"
      }
    ],
    "total": 1,
    "page": 1,
    "limit": 20,
    "total_pages": 1
  }
}
```

### 🔍 3. 特定の書籍を見る

```bash
curl http://localhost:8080/api/v1/books/1
```

### ✏️ 4. 書籍情報を更新する

```bash
curl -X PUT http://localhost:8080/api/v1/books/1 \
  -H "Content-Type: application/json" \
  -d '{
    "notes": "第3章まで読みました。とても分かりやすいです！",
    "tags": "プログラミング,Go言語,Web開発,お気に入り"
  }'
```

### 🗑️ 5. 書籍を削除する

```bash
curl -X DELETE http://localhost:8080/api/v1/books/1
```

## 📖 読書管理の操作

### 📚 1. 読書を開始する

```bash
curl -X POST http://localhost:8080/api/v1/books/1/start-reading
```

**期待される結果**：
```json
{
  "message": "読書を開始しました",
  "data": {
    "id": 1,
    "status": "reading",
    "start_read_date": "2024-01-15T12:30:00Z"
  }
}
```

### ✅ 2. 読書を完了する（評価付き）

```bash
curl -X POST http://localhost:8080/api/v1/books/1/finish-reading \
  -H "Content-Type: application/json" \
  -d '{
    "rating": 5
  }'
```

**期待される結果**：
```json
{
  "message": "読書を完了しました",
  "data": {
    "id": 1,
    "status": "completed",
    "end_read_date": "2024-01-20T18:00:00Z",
    "rating": 5
  }
}
```

### ✅ 3. 評価なしで読書を完了する

```bash
curl -X POST http://localhost:8080/api/v1/books/1/finish-reading
```

## 🔍 検索・フィルタリングの例

### 📚 1. 読書中の本だけを表示

```bash
curl "http://localhost:8080/api/v1/books?status=reading"
```

### 👤 2. 特定の著者の本を検索

```bash
curl "http://localhost:8080/api/v1/books?author=田中太郎"
```

### 🏷️ 3. タグで絞り込み

```bash
curl "http://localhost:8080/api/v1/books?tag=プログラミング"
```

### 🔍 4. タイトルや著者名で部分検索

```bash
curl "http://localhost:8080/api/v1/books?search=Go言語"
```

### ⭐ 5. 高評価（5点）の本だけを表示

```bash
curl "http://localhost:8080/api/v1/books?rating=5"
```

### 📄 6. ページングを使った表示

```bash
# 2ページ目を10件ずつ表示
curl "http://localhost:8080/api/v1/books?page=2&limit=10"
```

### 🔍 7. 複数条件を組み合わせた検索

```bash
# 読了済みで、5点評価で、プログラミング関連の本
curl "http://localhost:8080/api/v1/books?status=completed&rating=5&tag=プログラミング"
```

## 📊 統計情報の取得

### 📈 全体の統計を見る

```bash
curl http://localhost:8080/api/v1/statistics
```

**期待される結果**：
```json
{
  "message": "",
  "data": {
    "total_books": 50,
    "not_started_books": 20,
    "reading_books": 3,
    "completed_books": 25,
    "dropped_books": 2,
    "total_spent": 156000,
    "average_rating": 4.2,
    "books_this_month": 5,
    "completed_this_month": 3
  }
}
```

## 🔧 開発・デバッグ用

### ✅ サーバーの動作確認

```bash
curl http://localhost:8080/api/v1/health
```

**期待される結果**：
```json
{
  "message": "サービスは正常に動作しています",
  "data": {
    "status": "healthy",
    "service": "book-manager"
  }
}
```

## 🚨 エラーの例

### ❌ 1. 必須項目が足りない場合

```bash
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "タイトルのみ"
  }'
```

**エラー結果**：
```json
{
  "error": "入力データが無効です",
  "message": "Author is required"
}
```

### ❌ 2. 存在しない書籍にアクセス

```bash
curl http://localhost:8080/api/v1/books/999
```

**エラー結果**：
```json
{
  "error": "書籍が見つかりません",
  "message": "ID 999 の書籍が見つかりません"
}
```

### ❌ 3. 不正な評価値

```bash
curl -X POST http://localhost:8080/api/v1/books/1/finish-reading \
  -H "Content-Type: application/json" \
  -d '{
    "rating": 10
  }'
```

**エラー結果**：
```json
{
  "error": "読書完了に失敗しました",
  "message": "評価は1-5の範囲で入力してください: 10"
}
```

## 💡 実用的な使用例

### 📚 シナリオ1: 新しい本を買って読み始める

```bash
# 1. 本を登録
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Clean Code",
    "author": "Robert C. Martin",
    "purchase_date": "2024-01-15T00:00:00Z",
    "purchase_price": 4500,
    "tags": "プログラミング,設計"
  }'

# 2. 読書開始
curl -X POST http://localhost:8080/api/v1/books/1/start-reading

# 3. メモを追加
curl -X PUT http://localhost:8080/api/v1/books/1 \
  -H "Content-Type: application/json" \
  -d '{
    "notes": "命名規則について非常に参考になる"
  }'

# 4. 読書完了＆評価
curl -X POST http://localhost:8080/api/v1/books/1/finish-reading \
  -H "Content-Type: application/json" \
  -d '{"rating": 5}'
```

### 📊 シナリオ2: 読書習慣をチェック

```bash
# 1. 今月の統計を確認
curl http://localhost:8080/api/v1/statistics

# 2. 現在読書中の本を確認
curl "http://localhost:8080/api/v1/books?status=reading"

# 3. 高評価の本を振り返り
curl "http://localhost:8080/api/v1/books?rating=5&status=completed"
```

## 🛠️ 開発者向けテスト用データ

### 一括でテストデータを作成

```bash
# 複数の書籍を順番に登録
for i in {1..5}; do
  curl -X POST http://localhost:8080/api/v1/books \
    -H "Content-Type: application/json" \
    -d "{
      \"title\": \"テスト書籍 $i\",
      \"author\": \"テスト著者 $i\",
      \"purchase_date\": \"2024-01-0${i}T00:00:00Z\",
      \"purchase_price\": $((i * 1000))
    }"
done
```

このAPIリファレンスを使って、実際にアプリケーションの動作を確認してみてください！🚀
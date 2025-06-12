# 書籍管理アプリ

個人の書籍購入・読書管理を行うGo言語製のREST APIアプリケーションです。

## 機能

- 📚 書籍の登録・編集・削除
- 📖 読書ステータスの管理（未読・読書中・読了・中断）
- ⭐ 書籍の評価（1-5点）
- 🏷️ タグによる分類
- 📊 読書統計の表示
- 🔍 書籍の検索・フィルタリング

## 技術スタック

- **言語**: Go 1.21+
- **フレームワーク**: Gorilla Mux
- **データベース**: SQLite
- **アーキテクチャ**: クリーンアーキテクチャ

## セットアップ

### 前提条件

- Go 1.21以上
- Git

### インストール

```bash
# リポジトリをクローン
git clone <repository-url>
cd ccode-sample

# 依存関係をインストール
go mod download

# アプリケーションを実行
go run cmd/main.go
```

### アクセス方法

```bash
# WebUI（ブラウザで書籍管理）
http://localhost:8080

# API直接アクセス
http://localhost:8080/api/v1
```

### 環境変数

| 環境変数 | デフォルト値 | 説明 |
|---------|------------|------|
| `PORT` | 8080 | サーバーポート |
| `DB_PATH` | ./books.db | SQLiteデータベースファイルパス |

## API エンドポイント

### 書籍管理

#### 書籍を作成
```bash
POST /api/v1/books
Content-Type: application/json

{
  "title": "Go言語プログラミング",
  "author": "山田太郎",
  "isbn": "978-4-123-45678-9",
  "publisher": "技術出版社",
  "published_date": "2023-01-15T00:00:00Z",
  "purchase_date": "2023-02-01T00:00:00Z",
  "purchase_price": 3000,
  "tags": "プログラミング,Go言語",
  "notes": "基礎から学べる良い本"
}
```

#### 書籍一覧を取得
```bash
GET /api/v1/books?page=1&limit=20&status=not_started&search=Go
```

クエリパラメータ:
- `page`: ページ番号（デフォルト: 1）
- `limit`: 1ページあたりの件数（デフォルト: 20、最大: 100）
- `status`: 読書ステータス（`not_started`, `reading`, `completed`, `dropped`）
- `author`: 著者名での絞り込み
- `publisher`: 出版社での絞り込み
- `tag`: タグでの絞り込み
- `rating`: 評価での絞り込み（1-5）
- `search`: タイトル・著者名での部分一致検索

#### 書籍詳細を取得
```bash
GET /api/v1/books/{id}
```

#### 書籍を更新
```bash
PUT /api/v1/books/{id}
Content-Type: application/json

{
  "status": "reading",
  "notes": "読み始めました"
}
```

#### 書籍を削除
```bash
DELETE /api/v1/books/{id}
```

### 読書管理

#### 読書を開始
```bash
POST /api/v1/books/{id}/start-reading
```

#### 読書を完了
```bash
POST /api/v1/books/{id}/finish-reading
Content-Type: application/json

{
  "rating": 5
}
```

### 統計情報

#### 統計情報を取得
```bash
GET /api/v1/statistics
```

レスポンス例:
```json
{
  "message": "",
  "data": {
    "total_books": 150,
    "not_started_books": 80,
    "reading_books": 5,
    "completed_books": 60,
    "dropped_books": 5,
    "total_spent": 450000,
    "average_rating": 4.2,
    "books_this_month": 8,
    "completed_this_month": 3
  }
}
```

### その他

#### ヘルスチェック
```bash
GET /api/v1/health
```

## データモデル

### 書籍（Book）

| フィールド | 型 | 説明 |
|-----------|---|------|
| id | int | 書籍ID |
| title | string | タイトル |
| author | string | 著者 |
| isbn | string | ISBN |
| publisher | string | 出版社 |
| published_date | *time.Time | 出版日 |
| purchase_date | time.Time | 購入日 |
| purchase_price | int | 購入価格（円） |
| status | ReadingStatus | 読書ステータス |
| start_read_date | *time.Time | 読書開始日 |
| end_read_date | *time.Time | 読書終了日 |
| rating | *int | 評価（1-5点） |
| notes | string | メモ |
| tags | string | タグ（カンマ区切り） |
| created_at | time.Time | 作成日時 |
| updated_at | time.Time | 更新日時 |

### 読書ステータス（ReadingStatus）

- `not_started`: 未読
- `reading`: 読書中
- `completed`: 読了
- `dropped`: 中断

## 開発

### プロジェクト構造

```
ccode-sample/
├── cmd/                    # エントリーポイント
│   └── main.go
├── internal/               # 内部パッケージ
│   ├── model/              # データモデル
│   ├── repository/         # データアクセス層
│   ├── usecase/           # ビジネスロジック層
│   ├── handler/           # プレゼンテーション層
│   └── database/          # データベース設定
├── pkg/                   # 外部公開パッケージ
├── go.mod
├── go.sum
└── README.md
```

### ビルド
```bash
go build -o book-manager cmd/main.go
```

### テスト実行
```bash
go test ./...
```

### データベースの初期化
アプリケーション起動時に自動的にSQLiteデータベースが作成され、必要なテーブルがマイグレーションされます。

## 使用例

### 基本的な使用フロー

1. **書籍を登録**
```bash
curl -X POST http://localhost:8080/api/v1/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Clean Code",
    "author": "Robert C. Martin",
    "purchase_date": "2023-12-01T00:00:00Z",
    "purchase_price": 4000,
    "tags": "プログラミング,品質"
  }'
```

2. **読書を開始**
```bash
curl -X POST http://localhost:8080/api/v1/books/1/start-reading
```

3. **読書を完了**
```bash
curl -X POST http://localhost:8080/api/v1/books/1/finish-reading \
  -H "Content-Type: application/json" \
  -d '{"rating": 5}'
```

4. **統計情報を確認**
```bash
curl http://localhost:8080/api/v1/statistics
```

## ライセンス

MIT License

## 貢献

プルリクエストやイシューの報告を歓迎します。
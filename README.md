# 📚 書籍管理アプリ【Go/Web初心者向け学習プロジェクト】

このプロジェクトは、**Go言語とWeb開発を学びたい初心者**のための教材です。  
実際に動く書籍管理アプリケーションを通して、以下のことを学べます：

- 📖 Go言語の基本的な書き方
- 🌐 Webアプリケーションの作り方
- 🗄️ データベースの使い方
- 📱 API（アプリケーション間の連携）の作り方

## ✨ このアプリでできること

- 📚 **書籍の登録・編集・削除**：お気に入りの本を管理
- 📖 **読書ステータスの管理**：未読・読書中・読了・中断の状態を管理
- ⭐ **書籍の評価**：1-5点で本の評価をつける
- 🏷️ **タグによる分類**：ジャンルなどで本を分類
- 📊 **読書統計の表示**：どれくらい本を読んだかグラフで確認
- 🔍 **書籍の検索・フィルタリング**：本を簡単に見つける

## 🛠️ 使用技術（初心者向け解説）

- **Go言語**：Googleが作ったプログラミング言語。シンプルで学びやすい
- **SQLite**：軽量なデータベース。ファイル1つで動作する
- **HTML/CSS/JavaScript**：Webページを作るための基本技術
- **REST API**：アプリケーション間でデータをやり取りする仕組み

## 🚀 すぐに始める方法

### 1️⃣ 必要なソフトウェア

- **Go言語**：プログラムを実行するために必要
  - [公式サイト](https://golang.org/dl/)からダウンロード
  - バージョン1.21以上が必要
- **Git**：コードをダウンロードするために必要
  - [公式サイト](https://git-scm.com/)からダウンロード

### 2️⃣ アプリをダウンロードして実行

```bash
# 1. このプロジェクトをダウンロード
git clone https://github.com/taaag51/book-manager.git
cd book-manager

# 2. 必要なライブラリをインストール
go mod download

# 3. アプリケーションを起動
go run cmd/main.go
```

### 3️⃣ ブラウザで確認

アプリが起動したら、ブラウザで以下のURLを開いてください：

- **メインアプリ**：http://localhost:8080
- **API直接確認**：http://localhost:8080/api/v1/health

### 🔧 設定変更（上級者向け）

環境変数で設定を変更できます：

| 設定項目 | 環境変数名 | デフォルト値 | 説明 |
|---------|-----------|------------|------|
| ポート番号 | `PORT` | 8080 | アプリが使うポート番号 |
| データベースファイル | `DB_PATH` | ./books.db | データが保存される場所 |

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

## 📖 学習リソース

初心者の方は以下の順番で学習することをお勧めします：

### 🌟 基礎学習
1. **[Go言語超入門](docs/learning-go-web.md)** - Go言語とWeb開発の基礎
2. **[簡単Go講座](examples/simple-go-tutorial.md)** - このアプリで使われているGo言語の機能
3. **[API使用例集](examples/api-examples.md)** - 実際のAPI操作方法

### 🔍 コードリーディング
4. **[コードの流れ追跡法](docs/code-flow-tutorial.md)** - ユーザー操作からコード実行までの一気通貫理解
5. **[ステップ・バイ・ステップ解読](examples/step-by-step-code-reading.md)** - 実際のコードを一行ずつ解説

### 💪 実践・応用
6. **[実習課題](examples/exercise-beginner.md)** - 段階的な課題で実力アップ
7. **[デバッグガイド](docs/debug-guide.md)** - トラブルシューティングと問題解決

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
├── docs/                  # 学習用ドキュメント
├── examples/              # サンプルコードと実習課題
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
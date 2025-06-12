# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 重要な指示

**必ず日本語で回答してください。** このプロジェクトで作業する際は、すべての説明、コメント、エラーメッセージを日本語で提供してください。

## プロジェクト概要

個人向け書籍管理システム（REST API + Web UI）。Go言語とクリーンアーキテクチャで実装された、書籍の購入・読書進捗管理アプリケーション。WebUIとREST APIの両方を提供。

## 主要コマンド

### 開発・実行
- `go run cmd/main.go` - アプリケーション実行（ポート8080で起動）
- `go build -o book-manager cmd/main.go` - バイナリビルド
- `go mod download` - 依存関係の取得

### テスト・検証
- `go test ./...` - 全テスト実行
- `go test ./internal/model` - 特定パッケージのテスト実行
- `curl http://localhost:8080/api/v1/health` - ヘルスチェック

### WebUI アクセス
- `http://localhost:8080` - ブラウザでWebUI表示
- `http://localhost:8080/api/v1` - REST APIベースURL

### データベース
- データベースファイル: `./books.db` (SQLite)
- マイグレーション: アプリ起動時に自動実行

## アーキテクチャ

### クリーンアーキテクチャ（4層構造）
1. **Model層** (`internal/model/`) - データ構造定義、バリデーションルール
2. **Repository層** (`internal/repository/`) - データアクセス、SQLクエリ実装
3. **Usecase層** (`internal/usecase/`) - ビジネスロジック、統計処理
4. **Handler層** (`internal/handler/`) - HTTP処理、REST API エンドポイント

### 依存関係の流れ
```
Handler → Usecase → Repository → Database
  ↓         ↓         ↓
Model   ←  Model   ←  Model
```

### 主要コンポーネント
- **Database**: SQLite with embedded migration
- **Router**: Gorilla Mux with CORS middleware
- **Models**: Book entity with reading status lifecycle
- **Validation**: go-playground/validator for request validation

## データモデルの特徴

### Book Entity
- **読書ステータス**: `not_started` → `reading` → `completed`/`dropped`
- **自動日付管理**: ステータス変更時に読書開始・終了日を自動設定
- **統計計算**: 月次統計、平均評価、総支出額の自動計算

### Repository Pattern
- インターフェース定義による疎結合
- 動的クエリ構築によるフィルタリング機能
- ページネーション対応

## 環境設定

| 環境変数 | デフォルト | 説明 |
|---------|-----------|------|
| `PORT` | 8080 | HTTPサーバーポート |
| `DB_PATH` | ./books.db | SQLiteデータベースファイルパス |

## API設計パターン

### RESTful エンドポイント
- **書籍CRUD**: `/api/v1/books`
- **読書管理**: `/api/v1/books/{id}/start-reading`, `/api/v1/books/{id}/finish-reading`
- **統計情報**: `/api/v1/statistics`

### レスポンス形式
- 成功: `{"message": "", "data": {...}}`
- エラー: `{"error": "", "message": ""}`

### フィルタリング
複数条件での書籍検索をサポート（著者、出版社、タグ、ステータス、評価、全文検索）

## 開発時の注意点

### データベースマイグレーション
- `internal/database/migration.sql` に全スキーマ定義
- アプリ起動時に自動実行（冪等性保証）

### バリデーション
- 作成時: title, author, purchase_date は必須
- 評価: 1-5の範囲チェック
- 日付: 購入日の未来日チェック

### ログ・監視
- リクエストログ自動出力
- グレースフルシャットダウン対応
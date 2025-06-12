package database

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migration.sql
var migrationSQL embed.FS

// DB はデータベース接続を管理する構造体
type DB struct {
	*sql.DB
}

// NewDB は新しいデータベース接続を作成する
func NewDB(dataSourceName string) (*DB, error) {
	// SQLiteデータベースファイルのディレクトリを作成
	dir := filepath.Dir(dataSourceName)
	if dir != "." && dir != "" {
		// ディレクトリが存在しない場合は作成する（実際の運用では適切な権限設定が必要）
	}

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("データベースのオープンに失敗しました: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("データベースへの接続に失敗しました: %w", err)
	}

	return &DB{db}, nil
}

// Migrate はデータベースマイグレーションを実行する
func (db *DB) Migrate() error {
	migrationContent, err := migrationSQL.ReadFile("migration.sql")
	if err != nil {
		return fmt.Errorf("マイグレーションファイルの読み込みに失敗しました: %w", err)
	}

	if _, err := db.Exec(string(migrationContent)); err != nil {
		return fmt.Errorf("マイグレーションの実行に失敗しました: %w", err)
	}

	return nil
}

// Close はデータベース接続を閉じる
func (db *DB) Close() error {
	return db.DB.Close()
}
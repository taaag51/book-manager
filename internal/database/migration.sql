-- 書籍管理アプリ用のSQLiteデータベーススキーマ

-- 書籍テーブル
CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    isbn TEXT,
    publisher TEXT,
    published_date DATE,
    purchase_date DATE NOT NULL,
    purchase_price INTEGER DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'not_started' CHECK (status IN ('not_started', 'reading', 'completed', 'dropped')),
    start_read_date DATE,
    end_read_date DATE,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    notes TEXT,
    tags TEXT, -- カンマ区切りのタグ
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- インデックス作成
CREATE INDEX IF NOT EXISTS idx_books_status ON books(status);
CREATE INDEX IF NOT EXISTS idx_books_author ON books(author);
CREATE INDEX IF NOT EXISTS idx_books_publisher ON books(publisher);
CREATE INDEX IF NOT EXISTS idx_books_purchase_date ON books(purchase_date);
CREATE INDEX IF NOT EXISTS idx_books_rating ON books(rating);

-- 更新日時の自動更新用トリガー
CREATE TRIGGER IF NOT EXISTS update_books_updated_at
    AFTER UPDATE ON books
    FOR EACH ROW
BEGIN
    UPDATE books SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
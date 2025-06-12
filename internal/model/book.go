package model

import (
	"time"
)

// ReadingStatus は読書の状況を表す列挙型
type ReadingStatus string

const (
	StatusNotStarted ReadingStatus = "not_started" // 未読
	StatusReading    ReadingStatus = "reading"     // 読書中
	StatusCompleted  ReadingStatus = "completed"   // 読了
	StatusDropped    ReadingStatus = "dropped"     // 中断
)

// Book は書籍情報を表すモデル
type Book struct {
	ID            int           `json:"id" db:"id"`
	Title         string        `json:"title" db:"title"`
	Author        string        `json:"author" db:"author"`
	ISBN          string        `json:"isbn" db:"isbn"`
	Publisher     string        `json:"publisher" db:"publisher"`
	PublishedDate *time.Time    `json:"published_date" db:"published_date"`
	PurchaseDate  time.Time     `json:"purchase_date" db:"purchase_date"`
	PurchasePrice int           `json:"purchase_price" db:"purchase_price"`
	Status        ReadingStatus `json:"status" db:"status"`
	StartReadDate *time.Time    `json:"start_read_date" db:"start_read_date"`
	EndReadDate   *time.Time    `json:"end_read_date" db:"end_read_date"`
	Rating        *int          `json:"rating" db:"rating"` // 1-5点評価
	Notes         string        `json:"notes" db:"notes"`
	Tags          string        `json:"tags" db:"tags"` // カンマ区切りのタグ
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}

// CreateBookRequest は書籍作成時のリクエスト構造体
type CreateBookRequest struct {
	Title         string     `json:"title" validate:"required"`
	Author        string     `json:"author" validate:"required"`
	ISBN          string     `json:"isbn"`
	Publisher     string     `json:"publisher"`
	PublishedDate *time.Time `json:"published_date"`
	PurchaseDate  time.Time  `json:"purchase_date" validate:"required"`
	PurchasePrice int        `json:"purchase_price"`
	Tags          string     `json:"tags"`
	Notes         string     `json:"notes"`
}

// UpdateBookRequest は書籍更新時のリクエスト構造体
type UpdateBookRequest struct {
	Title         *string        `json:"title"`
	Author        *string        `json:"author"`
	ISBN          *string        `json:"isbn"`
	Publisher     *string        `json:"publisher"`
	PublishedDate *time.Time     `json:"published_date"`
	PurchasePrice *int           `json:"purchase_price"`
	Status        *ReadingStatus `json:"status"`
	StartReadDate *time.Time     `json:"start_read_date"`
	EndReadDate   *time.Time     `json:"end_read_date"`
	Rating        *int           `json:"rating"`
	Notes         *string        `json:"notes"`
	Tags          *string        `json:"tags"`
}

// BookFilter は書籍検索用のフィルター構造体
type BookFilter struct {
	Status    *ReadingStatus `json:"status"`
	Author    *string        `json:"author"`
	Publisher *string        `json:"publisher"`
	Tag       *string        `json:"tag"`
	Rating    *int           `json:"rating"`
	Search    *string        `json:"search"` // タイトル・著者の部分一致検索
}
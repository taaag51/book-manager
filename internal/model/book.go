// modelパッケージ：データの構造（型）を定義するファイル
// 書籍の情報やリクエスト形式など、データの「型」を決める
package model

import (
	"time"  // 時間関連の型（time.Time）を使うため
)

// ReadingStatus は読書の状況を表す列挙型（決められた値のみ使える型）
// string型をベースにして、決められた値だけを使えるようにする
type ReadingStatus string

// 読書ステータスの定数定義
// const：変更できない値（定数）を定義
const (
	StatusNotStarted ReadingStatus = "not_started" // 未読（まだ読んでいない）
	StatusReading    ReadingStatus = "reading"     // 読書中（現在読んでいる）
	StatusCompleted  ReadingStatus = "completed"   // 読了（読み終わった）
	StatusDropped    ReadingStatus = "dropped"     // 中断（途中でやめた）
)

// Book は書籍情報を表すモデル（データ構造）
// struct：複数のデータをまとめて一つの型にする仕組み
// `json:"xxx"` と `db:"xxx"`：JSON形式とデータベースでの項目名を指定
type Book struct {
	ID            int           `json:"id" db:"id"`                         // 書籍の一意なID番号
	Title         string        `json:"title" db:"title"`                   // 書籍のタイトル
	Author        string        `json:"author" db:"author"`                 // 著者名
	ISBN          string        `json:"isbn" db:"isbn"`                     // ISBN番号（本の識別番号）
	Publisher     string        `json:"publisher" db:"publisher"`           // 出版社名
	PublishedDate *time.Time    `json:"published_date" db:"published_date"` // 出版日（*は値がnullの可能性があることを示す）
	PurchaseDate  time.Time     `json:"purchase_date" db:"purchase_date"`   // 購入日
	PurchasePrice int           `json:"purchase_price" db:"purchase_price"` // 購入価格（円）
	Status        ReadingStatus `json:"status" db:"status"`                 // 読書ステータス
	StartReadDate *time.Time    `json:"start_read_date" db:"start_read_date"` // 読書開始日（nullable）
	EndReadDate   *time.Time    `json:"end_read_date" db:"end_read_date"`   // 読書終了日（nullable）
	Rating        *int          `json:"rating" db:"rating"`                 // 評価（1-5点、nullable）
	Notes         string        `json:"notes" db:"notes"`                   // メモ・感想
	Tags          string        `json:"tags" db:"tags"`                     // タグ（カンマ区切り文字列）
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`         // 作成日時
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`         // 更新日時
}

// CreateBookRequest は書籍作成時のリクエスト構造体
// APIで新しい書籍を作成する時に送信するデータの形式
// `validate:"required"`：この項目は必須入力であることを示す
type CreateBookRequest struct {
	Title         string     `json:"title" validate:"required"`         // タイトル（必須）
	Author        string     `json:"author" validate:"required"`        // 著者（必須）
	ISBN          string     `json:"isbn"`                              // ISBN番号（任意）
	Publisher     string     `json:"publisher"`                         // 出版社（任意）
	PublishedDate *time.Time `json:"published_date"`                   // 出版日（任意、nullの可能性あり）
	PurchaseDate  time.Time  `json:"purchase_date" validate:"required"` // 購入日（必須）
	PurchasePrice int        `json:"purchase_price"`                   // 購入価格（任意）
	Tags          string     `json:"tags"`                              // タグ（任意）
	Notes         string     `json:"notes"`                             // メモ（任意）
}

// UpdateBookRequest は書籍更新時のリクエスト構造体
// APIで既存の書籍情報を更新する時に送信するデータの形式
// 全ての項目が*（ポインタ）になっているのは、更新しない項目はnullを送るため
type UpdateBookRequest struct {
	Title         *string        `json:"title"`          // タイトル（更新する場合のみ）
	Author        *string        `json:"author"`         // 著者（更新する場合のみ）
	ISBN          *string        `json:"isbn"`           // ISBN番号（更新する場合のみ）
	Publisher     *string        `json:"publisher"`      // 出版社（更新する場合のみ）
	PublishedDate *time.Time     `json:"published_date"` // 出版日（更新する場合のみ）
	PurchasePrice *int           `json:"purchase_price"` // 購入価格（更新する場合のみ）
	Status        *ReadingStatus `json:"status"`         // 読書ステータス（更新する場合のみ）
	StartReadDate *time.Time     `json:"start_read_date"` // 読書開始日（更新する場合のみ）
	EndReadDate   *time.Time     `json:"end_read_date"`  // 読書終了日（更新する場合のみ）
	Rating        *int           `json:"rating"`         // 評価（更新する場合のみ）
	Notes         *string        `json:"notes"`          // メモ（更新する場合のみ）
	Tags          *string        `json:"tags"`           // タグ（更新する場合のみ）
}

// BookFilter は書籍検索用のフィルター構造体
// 書籍一覧を取得する時の検索・絞り込み条件を指定する形式
type BookFilter struct {
	Status    *ReadingStatus `json:"status"`    // 読書ステータスで絞り込み
	Author    *string        `json:"author"`    // 著者名で絞り込み
	Publisher *string        `json:"publisher"` // 出版社で絞り込み
	Tag       *string        `json:"tag"`       // タグで絞り込み
	Rating    *int           `json:"rating"`    // 評価で絞り込み
	Search    *string        `json:"search"`    // タイトル・著者の部分一致検索
}
package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"book-manager/internal/database"
	"book-manager/internal/model"
)

// BookRepository は書籍データの永続化を担当するインターフェース
type BookRepository interface {
	Create(book *model.CreateBookRequest) (*model.Book, error)
	GetByID(id int) (*model.Book, error)
	List(filter *model.BookFilter, limit, offset int) ([]*model.Book, error)
	Update(id int, book *model.UpdateBookRequest) (*model.Book, error)
	Delete(id int) error
	Count(filter *model.BookFilter) (int, error)
}

// bookRepository はBookRepositoryの実装
type bookRepository struct {
	db *database.DB
}

// NewBookRepository は新しいBookRepositoryを作成する
func NewBookRepository(db *database.DB) BookRepository {
	return &bookRepository{db: db}
}

// Create は新しい書籍を作成する
func (r *bookRepository) Create(req *model.CreateBookRequest) (*model.Book, error) {
	query := `
		INSERT INTO books (title, author, isbn, publisher, published_date, purchase_date, purchase_price, tags, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		req.Title,
		req.Author,
		req.ISBN,
		req.Publisher,
		req.PublishedDate,
		req.PurchaseDate,
		req.PurchasePrice,
		req.Tags,
		req.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("書籍の作成に失敗しました: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("書籍IDの取得に失敗しました: %w", err)
	}

	return r.GetByID(int(id))
}

// GetByID は指定されたIDの書籍を取得する
func (r *bookRepository) GetByID(id int) (*model.Book, error) {
	query := `
		SELECT id, title, author, isbn, publisher, published_date, purchase_date, 
		       purchase_price, status, start_read_date, end_read_date, rating, 
		       notes, tags, created_at, updated_at
		FROM books 
		WHERE id = ?
	`

	book := &model.Book{}
	row := r.db.QueryRow(query, id)

	err := row.Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.ISBN,
		&book.Publisher,
		&book.PublishedDate,
		&book.PurchaseDate,
		&book.PurchasePrice,
		&book.Status,
		&book.StartReadDate,
		&book.EndReadDate,
		&book.Rating,
		&book.Notes,
		&book.Tags,
		&book.CreatedAt,
		&book.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ID %d の書籍が見つかりません", id)
		}
		return nil, fmt.Errorf("書籍の取得に失敗しました: %w", err)
	}

	return book, nil
}

// List はフィルター条件に基づいて書籍一覧を取得する
func (r *bookRepository) List(filter *model.BookFilter, limit, offset int) ([]*model.Book, error) {
	query := "SELECT id, title, author, isbn, publisher, published_date, purchase_date, purchase_price, status, start_read_date, end_read_date, rating, notes, tags, created_at, updated_at FROM books"
	args := []interface{}{}
	conditions := []string{}

	// フィルター条件を構築
	if filter != nil {
		if filter.Status != nil {
			conditions = append(conditions, "status = ?")
			args = append(args, *filter.Status)
		}
		if filter.Author != nil {
			conditions = append(conditions, "author = ?")
			args = append(args, *filter.Author)
		}
		if filter.Publisher != nil {
			conditions = append(conditions, "publisher = ?")
			args = append(args, *filter.Publisher)
		}
		if filter.Rating != nil {
			conditions = append(conditions, "rating = ?")
			args = append(args, *filter.Rating)
		}
		if filter.Tag != nil {
			conditions = append(conditions, "tags LIKE ?")
			args = append(args, "%"+*filter.Tag+"%")
		}
		if filter.Search != nil {
			conditions = append(conditions, "(title LIKE ? OR author LIKE ?)")
			searchTerm := "%" + *filter.Search + "%"
			args = append(args, searchTerm, searchTerm)
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
		if offset > 0 {
			query += " OFFSET ?"
			args = append(args, offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("書籍一覧の取得に失敗しました: %w", err)
	}
	defer rows.Close()

	books := []*model.Book{}
	for rows.Next() {
		book := &model.Book{}
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.ISBN,
			&book.Publisher,
			&book.PublishedDate,
			&book.PurchaseDate,
			&book.PurchasePrice,
			&book.Status,
			&book.StartReadDate,
			&book.EndReadDate,
			&book.Rating,
			&book.Notes,
			&book.Tags,
			&book.CreatedAt,
			&book.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("書籍データの読み込みに失敗しました: %w", err)
		}
		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("書籍一覧の処理中にエラーが発生しました: %w", err)
	}

	return books, nil
}

// Update は書籍情報を更新する
func (r *bookRepository) Update(id int, req *model.UpdateBookRequest) (*model.Book, error) {
	setParts := []string{}
	args := []interface{}{}

	if req.Title != nil {
		setParts = append(setParts, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Author != nil {
		setParts = append(setParts, "author = ?")
		args = append(args, *req.Author)
	}
	if req.ISBN != nil {
		setParts = append(setParts, "isbn = ?")
		args = append(args, *req.ISBN)
	}
	if req.Publisher != nil {
		setParts = append(setParts, "publisher = ?")
		args = append(args, *req.Publisher)
	}
	if req.PublishedDate != nil {
		setParts = append(setParts, "published_date = ?")
		args = append(args, *req.PublishedDate)
	}
	if req.PurchasePrice != nil {
		setParts = append(setParts, "purchase_price = ?")
		args = append(args, *req.PurchasePrice)
	}
	if req.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *req.Status)
		
		// ステータスに応じて読書開始・終了日を自動設定
		now := time.Now()
		if *req.Status == model.StatusReading && req.StartReadDate == nil {
			setParts = append(setParts, "start_read_date = ?")
			args = append(args, now)
		}
		if (*req.Status == model.StatusCompleted || *req.Status == model.StatusDropped) && req.EndReadDate == nil {
			setParts = append(setParts, "end_read_date = ?")
			args = append(args, now)
		}
	}
	if req.StartReadDate != nil {
		setParts = append(setParts, "start_read_date = ?")
		args = append(args, *req.StartReadDate)
	}
	if req.EndReadDate != nil {
		setParts = append(setParts, "end_read_date = ?")
		args = append(args, *req.EndReadDate)
	}
	if req.Rating != nil {
		setParts = append(setParts, "rating = ?")
		args = append(args, *req.Rating)
	}
	if req.Notes != nil {
		setParts = append(setParts, "notes = ?")
		args = append(args, *req.Notes)
	}
	if req.Tags != nil {
		setParts = append(setParts, "tags = ?")
		args = append(args, *req.Tags)
	}

	if len(setParts) == 0 {
		return r.GetByID(id)
	}

	query := "UPDATE books SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("書籍の更新に失敗しました: %w", err)
	}

	return r.GetByID(id)
}

// Delete は書籍を削除する
func (r *bookRepository) Delete(id int) error {
	query := "DELETE FROM books WHERE id = ?"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("書籍の削除に失敗しました: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("削除結果の確認に失敗しました: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ID %d の書籍が見つかりません", id)
	}

	return nil
}

// Count はフィルター条件に一致する書籍数を取得する
func (r *bookRepository) Count(filter *model.BookFilter) (int, error) {
	query := "SELECT COUNT(*) FROM books"
	args := []interface{}{}
	conditions := []string{}

	if filter != nil {
		if filter.Status != nil {
			conditions = append(conditions, "status = ?")
			args = append(args, *filter.Status)
		}
		if filter.Author != nil {
			conditions = append(conditions, "author = ?")
			args = append(args, *filter.Author)
		}
		if filter.Publisher != nil {
			conditions = append(conditions, "publisher = ?")
			args = append(args, *filter.Publisher)
		}
		if filter.Rating != nil {
			conditions = append(conditions, "rating = ?")
			args = append(args, *filter.Rating)
		}
		if filter.Tag != nil {
			conditions = append(conditions, "tags LIKE ?")
			args = append(args, "%"+*filter.Tag+"%")
		}
		if filter.Search != nil {
			conditions = append(conditions, "(title LIKE ? OR author LIKE ?)")
			searchTerm := "%" + *filter.Search + "%"
			args = append(args, searchTerm, searchTerm)
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("書籍数の取得に失敗しました: %w", err)
	}

	return count, nil
}
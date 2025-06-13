// repositoryパッケージ：データベースとのやり取りを担当するファイル
// Repository層は「データアクセス層」とも呼ばれ、データベースの操作を専門に行う
package repository

// import：他のパッケージ（機能）を使うための宣言
import (
	"database/sql"                   // データベース操作の基本機能
	"fmt"                           // 文字列フォーマット（%vなどの置き換え）
	"strings"                       // 文字列操作（結合、分割など）
	"time"                          // 時間関連の処理

	"book-manager/internal/database" // 自作のデータベース接続機能
	"book-manager/internal/model"    // 自作のデータ構造定義
)

// BookRepository は書籍データの永続化を担当するインターフェース
// インターフェース：「こんな機能を持つ型」を定義する仕組み
// 永続化：データをデータベースに保存すること（プログラム終了後も残る）
type BookRepository interface {
	Create(book *model.CreateBookRequest) (*model.Book, error)   // 新しい書籍をデータベースに保存
	GetByID(id int) (*model.Book, error)                         // IDで書籍を1件取得
	List(filter *model.BookFilter, limit, offset int) ([]*model.Book, error) // 条件に合う書籍リストを取得
	Update(id int, book *model.UpdateBookRequest) (*model.Book, error)        // 書籍情報を更新
	Delete(id int) error                                         // 書籍を削除
	Count(filter *model.BookFilter) (int, error)                // 条件に合う書籍数をカウント
}

// bookRepository はBookRepositoryインターフェースの実装
// struct：複数のデータをまとめた構造体
// *database.DB：データベース接続を保持（*はポインタ型）
type bookRepository struct {
	db *database.DB // データベース接続オブジェクト
}

// NewBookRepository は新しいBookRepositoryを作成する関数
// コンストラクタ関数：新しいインスタンス（実体）を作る関数
// &：アドレス演算子（メモリ上の場所を示すポインタを作る）
func NewBookRepository(db *database.DB) BookRepository {
	return &bookRepository{db: db} // bookRepository構造体のポインタを返す
}

// Create は新しい書籍をデータベースに保存する関数
// (r *bookRepository)：レシーバー（この関数がどの型に属するかを示す）
// req *model.CreateBookRequest：作成用のリクエストデータ
// (*model.Book, error)：戻り値（作成された書籍データとエラー）
func (r *bookRepository) Create(req *model.CreateBookRequest) (*model.Book, error) {
	// query：SQL文（データベースに実行させる命令）
	// INSERT INTO：新しいデータを挿入するSQL命令
	// ?：プレースホルダー（後で実際の値に置き換えられる）
	query := `
		INSERT INTO books (title, author, isbn, publisher, published_date, purchase_date, purchase_price, tags, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// r.db.Exec()：SQLを実行する関数
	// プレースホルダー（?）に実際の値を順番に入れて実行
	result, err := r.db.Exec(query,
		req.Title,         // タイトル
		req.Author,        // 著者
		req.ISBN,          // ISBN
		req.Publisher,     // 出版社
		req.PublishedDate, // 出版日
		req.PurchaseDate,  // 購入日
		req.PurchasePrice, // 購入価格
		req.Tags,          // タグ
		req.Notes,         // メモ
	)
	// エラーハンドリング：エラーが発生した場合の処理
	if err != nil {
		return nil, fmt.Errorf("書籍の作成に失敗しました: %w", err)
	}

	// LastInsertId()：挿入されたデータの自動生成ID（主キー）を取得
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("書籍IDの取得に失敗しました: %w", err)
	}

	// 作成された書籍のデータを取得して返す
	// int(id)：int64型をint型に変換
	return r.GetByID(int(id))
}

// GetByID は指定されたIDの書籍を1件取得する関数
// SELECT：データベースからデータを取得するSQL命令
func (r *bookRepository) GetByID(id int) (*model.Book, error) {
	// SELECT文：booksテーブルから指定したカラム（列）のデータを取得
	// WHERE id = ?：IDが一致する行だけを取得する条件
	query := `
		SELECT id, title, author, isbn, publisher, published_date, purchase_date, 
		       purchase_price, status, start_read_date, end_read_date, rating, 
		       notes, tags, created_at, updated_at
		FROM books 
		WHERE id = ?
	`

	// &model.Book{}：空のBook構造体を作成（&でポインタにする）
	book := &model.Book{}
	// QueryRow()：1行だけを取得するSQL実行関数
	row := r.db.QueryRow(query, id)

	// Scan()：取得したデータを構造体の各フィールドに格納
	// &book.ID：bookのIDフィールドのアドレス（格納先を指定）
	err := row.Scan(
		&book.ID,            // 書籍ID
		&book.Title,         // タイトル
		&book.Author,        // 著者
		&book.ISBN,          // ISBN
		&book.Publisher,     // 出版社
		&book.PublishedDate, // 出版日
		&book.PurchaseDate,  // 購入日
		&book.PurchasePrice, // 購入価格
		&book.Status,        // 読書ステータス
		&book.StartReadDate, // 読書開始日
		&book.EndReadDate,   // 読書終了日
		&book.Rating,        // 評価
		&book.Notes,         // メモ
		&book.Tags,          // タグ
		&book.CreatedAt,     // 作成日時
		&book.UpdatedAt,     // 更新日時
	)

	// エラーハンドリング
	if err != nil {
		// sql.ErrNoRows：該当するデータが見つからない場合の特別なエラー
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ID %d の書籍が見つかりません", id)
		}
		return nil, fmt.Errorf("書籍の取得に失敗しました: %w", err)
	}

	// 正常終了：取得した書籍データとnilエラーを返す
	return book, nil
}

// List はフィルター条件に基づいて書籍一覧を取得する関数
// []*model.Book：Book構造体のポインタのスライス（配列）
// limit：最大取得件数、offset：何件目から取得するか（ページング用）
func (r *bookRepository) List(filter *model.BookFilter, limit, offset int) ([]*model.Book, error) {
	// 基本のSELECT文
	query := "SELECT id, title, author, isbn, publisher, published_date, purchase_date, purchase_price, status, start_read_date, end_read_date, rating, notes, tags, created_at, updated_at FROM books"
	// args：SQLのプレースホルダーに入れる値のスライス
	args := []interface{}{}
	// conditions：WHERE句の条件文のスライス
	conditions := []string{}

	// フィルター条件を動的に構築
	// 動的SQL：条件に応じてSQL文を組み立てる手法
	if filter != nil {
		// *filter.Status：ポインタの値を取得（*は逆参照演算子）
		if filter.Status != nil {
			conditions = append(conditions, "status = ?")   // 読書ステータスで絞り込み
			args = append(args, *filter.Status)
		}
		if filter.Author != nil {
			conditions = append(conditions, "author = ?")    // 著者名で絞り込み
			args = append(args, *filter.Author)
		}
		if filter.Publisher != nil {
			conditions = append(conditions, "publisher = ?") // 出版社で絞り込み
			args = append(args, *filter.Publisher)
		}
		if filter.Rating != nil {
			conditions = append(conditions, "rating = ?")    // 評価で絞り込み
			args = append(args, *filter.Rating)
		}
		if filter.Tag != nil {
			// LIKE：部分一致検索、%は任意の文字列を表すワイルドカード
			conditions = append(conditions, "tags LIKE ?")   // タグで部分一致検索
			args = append(args, "%"+*filter.Tag+"%")
		}
		if filter.Search != nil {
			// OR：複数条件のいずれかに一致
			conditions = append(conditions, "(title LIKE ? OR author LIKE ?)")
			searchTerm := "%" + *filter.Search + "%"  // 前後にワイルドカードを付加
			args = append(args, searchTerm, searchTerm) // タイトルと著者の両方に同じ条件
		}
	}

	// 条件がある場合はWHERE句を追加
	if len(conditions) > 0 {
		// strings.Join()：スライスを指定した文字で結合
		// AND：複数条件をすべて満たす場合
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// ORDER BY：結果の並び順を指定（created_at DESC = 作成日時の降順）
	query += " ORDER BY created_at DESC"

	// ページング処理（LIMIT：件数制限、OFFSET：開始位置）
	if limit > 0 {
		query += " LIMIT ?"    // 最大取得件数
		args = append(args, limit)
		if offset > 0 {
			query += " OFFSET ?" // 開始位置（スキップする件数）
			args = append(args, offset)
		}
	}

	// Query()：複数行を取得するSQL実行関数
	// args...：スライスを可変長引数として展開
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("書籍一覧の取得に失敗しました: %w", err)
	}
	// defer：関数終了時に必ず実行される処理（リソース解放）
	defer rows.Close()

	// 結果を格納するスライスを初期化
	books := []*model.Book{}
	// rows.Next()：次の行があるかチェック（forループで全行を処理）
	for rows.Next() {
		// 各行ごとに新しいBook構造体を作成
		book := &model.Book{}
		// 1行分のデータを構造体のフィールドに格納
		err := rows.Scan(
			&book.ID,            // ID
			&book.Title,         // タイトル
			&book.Author,        // 著者
			&book.ISBN,          // ISBN
			&book.Publisher,     // 出版社
			&book.PublishedDate, // 出版日
			&book.PurchaseDate,  // 購入日
			&book.PurchasePrice, // 購入価格
			&book.Status,        // 読書ステータス
			&book.StartReadDate, // 読書開始日
			&book.EndReadDate,   // 読書終了日
			&book.Rating,        // 評価
			&book.Notes,         // メモ
			&book.Tags,          // タグ
			&book.CreatedAt,     // 作成日時
			&book.UpdatedAt,     // 更新日時
		)
		if err != nil {
			return nil, fmt.Errorf("書籍データの読み込みに失敗しました: %w", err)
		}
		// スライスに書籍データを追加
		books = append(books, book)
	}

	// rows.Err()：ループ処理中にエラーが発生していないかチェック
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("書籍一覧の処理中にエラーが発生しました: %w", err)
	}

	// 正常終了：取得した書籍リストを返す
	return books, nil
}

// Update は書籍情報を更新する関数
// 更新するフィールドだけを動的にUPDATE文に含める
func (r *bookRepository) Update(id int, req *model.UpdateBookRequest) (*model.Book, error) {
	// setParts：UPDATE文のSET句の部分
	setParts := []string{}
	// args：プレースホルダーに入れる値
	args := []interface{}{}

	// 更新対象のフィールドがある場合のみSET句に追加
	// nil チェック：値が設定されているかを確認
	if req.Title != nil {
		setParts = append(setParts, "title = ?")  // タイトル更新
		args = append(args, *req.Title)
	}
	if req.Author != nil {
		setParts = append(setParts, "author = ?") // 著者更新
		args = append(args, *req.Author)
	}
	if req.ISBN != nil {
		setParts = append(setParts, "isbn = ?")   // ISBN更新
		args = append(args, *req.ISBN)
	}
	if req.Publisher != nil {
		setParts = append(setParts, "publisher = ?") // 出版社更新
		args = append(args, *req.Publisher)
	}
	if req.PublishedDate != nil {
		setParts = append(setParts, "published_date = ?") // 出版日更新
		args = append(args, *req.PublishedDate)
	}
	if req.PurchasePrice != nil {
		setParts = append(setParts, "purchase_price = ?") // 購入価格更新
		args = append(args, *req.PurchasePrice)
	}
	if req.Status != nil {
		setParts = append(setParts, "status = ?")  // 読書ステータス更新
		args = append(args, *req.Status)
		
		// ビジネスロジック：ステータスに応じて読書開始・終了日を自動設定
		now := time.Now()  // 現在時刻を取得
		// 読書中になった場合、開始日が未設定なら現在時刻を設定
		if *req.Status == model.StatusReading && req.StartReadDate == nil {
			setParts = append(setParts, "start_read_date = ?")
			args = append(args, now)
		}
		// 読了・中断になった場合、終了日が未設定なら現在時刻を設定
		if (*req.Status == model.StatusCompleted || *req.Status == model.StatusDropped) && req.EndReadDate == nil {
			setParts = append(setParts, "end_read_date = ?")
			args = append(args, now)
		}
	}
	if req.StartReadDate != nil {
		setParts = append(setParts, "start_read_date = ?") // 読書開始日更新
		args = append(args, *req.StartReadDate)
	}
	if req.EndReadDate != nil {
		setParts = append(setParts, "end_read_date = ?")   // 読書終了日更新
		args = append(args, *req.EndReadDate)
	}
	if req.Rating != nil {
		setParts = append(setParts, "rating = ?")          // 評価更新
		args = append(args, *req.Rating)
	}
	if req.Notes != nil {
		setParts = append(setParts, "notes = ?")           // メモ更新
		args = append(args, *req.Notes)
	}
	if req.Tags != nil {
		setParts = append(setParts, "tags = ?")            // タグ更新
		args = append(args, *req.Tags)
	}

	// 更新するフィールドがない場合は、現在のデータをそのまま返す
	if len(setParts) == 0 {
		return r.GetByID(id)
	}

	// UPDATE文を動的に構築
	// strings.Join()：SET句の各部分をカンマで結合
	query := "UPDATE books SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	args = append(args, id)  // WHERE句のIDをパラメータに追加

	// UPDATE文を実行
	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("書籍の更新に失敗しました: %w", err)
	}

	// 更新後のデータを取得して返す
	return r.GetByID(id)
}

// Delete は書籍をデータベースから削除する関数
func (r *bookRepository) Delete(id int) error {
	// DELETE文：指定したIDの書籍を削除
	query := "DELETE FROM books WHERE id = ?"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("書籍の削除に失敗しました: %w", err)
	}

	// RowsAffected()：実際に削除された行数を取得
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("削除結果の確認に失敗しました: %w", err)
	}

	// 削除された行数が0の場合、該当するIDの書籍が存在しなかった
	if rowsAffected == 0 {
		return fmt.Errorf("ID %d の書籍が見つかりません", id)
	}

	// 正常終了（エラーなし）
	return nil
}

// Count はフィルター条件に一致する書籍数を取得する関数
// ページング処理で「全何件中〇件目」を表示するために使用
func (r *bookRepository) Count(filter *model.BookFilter) (int, error) {
	// COUNT(*)：テーブルの行数を数えるSQL関数
	query := "SELECT COUNT(*) FROM books"
	args := []interface{}{}
	conditions := []string{}

	// Listメソッドと同じフィルター条件を適用
	// カウント対象を絞り込む
	if filter != nil {
		if filter.Status != nil {
			conditions = append(conditions, "status = ?")     // ステータス絞り込み
			args = append(args, *filter.Status)
		}
		if filter.Author != nil {
			conditions = append(conditions, "author = ?")      // 著者絞り込み
			args = append(args, *filter.Author)
		}
		if filter.Publisher != nil {
			conditions = append(conditions, "publisher = ?")   // 出版社絞り込み
			args = append(args, *filter.Publisher)
		}
		if filter.Rating != nil {
			conditions = append(conditions, "rating = ?")      // 評価絞り込み
			args = append(args, *filter.Rating)
		}
		if filter.Tag != nil {
			conditions = append(conditions, "tags LIKE ?")     // タグ部分一致
			args = append(args, "%"+*filter.Tag+"%")
		}
		if filter.Search != nil {
			conditions = append(conditions, "(title LIKE ? OR author LIKE ?)") // 全文検索
			searchTerm := "%" + *filter.Search + "%"
			args = append(args, searchTerm, searchTerm)
		}
	}

	// 条件がある場合はWHERE句を追加
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// カウント結果を格納する変数
	var count int
	// QueryRow()で1つの値（カウント数）を取得
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("書籍数の取得に失敗しました: %w", err)
	}

	// カウント数を返す
	return count, nil
}
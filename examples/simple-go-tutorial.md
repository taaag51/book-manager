# 🐹 超初心者向けGo言語ミニ講座

このファイルでは、Go言語を全く知らない人向けに、書籍管理アプリで使われているGo言語の基本を説明します。

## 🎯 Go言語って何？

Go言語は**Google**が作ったプログラミング言語です。特徴は：

- **シンプル**：覚えることが少なく、読みやすい
- **高速**：プログラムが速く動く
- **安全**：バグが起きにくい仕組みがある

## 📝 基本的な書き方

### 1. Hello World

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

**解説**：
- `package main`：このファイルがメインプログラムであることを宣言
- `import "fmt"`：文字表示機能を使うための準備
- `func main()`：プログラムの開始地点
- `fmt.Println()`：文字を画面に表示する命令

### 2. 変数の宣言

```go
// 基本的な変数の宣言
var name string = "田中太郎"  // 文字列
var age int = 25             // 整数
var price float64 = 1500.50  // 小数

// 短縮形（よく使う）
title := "Go言語入門"  // 型は自動判定される
count := 10
```

**解説**：
- `var`：変数を作る時のキーワード
- `string`：文字列の型
- `int`：整数の型
- `:=`：型を自動で判定してくれる短縮記法

### 3. 構造体（データをまとめる）

```go
// Book構造体の定義
type Book struct {
    Title  string  // タイトル
    Author string  // 著者
    Price  int     // 価格
}

// 構造体を使う
func main() {
    // 新しい本を作る
    book := Book{
        Title:  "Go言語入門",
        Author: "山田太郎",
        Price:  3000,
    }
    
    // 値を取り出す
    fmt.Println(book.Title)  // "Go言語入門"
    fmt.Println(book.Price)  // 3000
}
```

**解説**：
- `type`：新しい型を定義するキーワード
- `struct`：複数の値をまとめて管理する箱
- `book.Title`：構造体の中の値にアクセス

## 🔍 書籍管理アプリで使われている重要概念

### 1. ポインタ（*）

```go
// 普通の変数
var x int = 42

// ポインタ変数（xの住所を保存）
var p *int = &x

fmt.Println(x)   // 42（値そのもの）
fmt.Println(p)   // 0xc000014098（xの住所）
fmt.Println(*p)  // 42（住所の先にある値）
```

**実用例**：
```go
// なぜポインタが必要？ → 「値がない」ことを表現できる
type Book struct {
    Title  string  // 必須項目
    Rating *int    // 評価（未評価の場合はnil）
}

book := Book{
    Title:  "Go入門",
    Rating: nil,  // まだ評価していない
}

if book.Rating != nil {
    fmt.Println("評価:", *book.Rating)
} else {
    fmt.Println("未評価")
}
```

### 2. 関数とメソッド

```go
// 普通の関数
func add(a, b int) int {
    return a + b
}

// メソッド（構造体に紐づく関数）
type Calculator struct {
    result int
}

func (c *Calculator) Add(x int) {
    c.result += x
}

func (c *Calculator) GetResult() int {
    return c.result
}

// 使い方
calc := &Calculator{}
calc.Add(10)
calc.Add(5)
fmt.Println(calc.GetResult())  // 15
```

**書籍管理アプリでの例**：
```go
// BookRepositoryのCreateメソッド
func (r *bookRepository) Create(req *model.CreateBookRequest) (*model.Book, error) {
    // データベースに本を保存する処理
    return book, nil
}
```

### 3. インターフェース

```go
// 「こんな機能を持つもの」を定義
type Writer interface {
    Write(data string) error
}

// ファイルに書く構造体
type FileWriter struct {
    filename string
}

func (f *FileWriter) Write(data string) error {
    // ファイルに書く処理
    return nil
}

// データベースに書く構造体
type DatabaseWriter struct {
    connection string
}

func (d *DatabaseWriter) Write(data string) error {
    // データベースに書く処理
    return nil
}

// どちらでも使える関数
func saveData(w Writer, data string) {
    w.Write(data)  // FileWriterでもDatabaseWriterでもOK
}
```

**書籍管理アプリでの例**：
```go
// BookRepositoryインターフェース
type BookRepository interface {
    Create(book *model.CreateBookRequest) (*model.Book, error)
    GetByID(id int) (*model.Book, error)
}

// 実際の実装
type bookRepository struct {
    db *database.DB
}

func (r *bookRepository) Create(book *model.CreateBookRequest) (*model.Book, error) {
    // SQLiteに保存する実装
}
```

### 4. エラーハンドリング

```go
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("0で割ることはできません")
    }
    return a / b, nil
}

func main() {
    result, err := divide(10, 2)
    if err != nil {
        fmt.Println("エラー:", err)
        return
    }
    fmt.Println("結果:", result)
}
```

**重要なポイント**：
- Goでは例外（try-catch）がない
- エラーは戻り値として明示的に返す
- エラーチェックは必須

## 🏗️ 書籍管理アプリの構造を理解しよう

### 1. パッケージ構成

```
internal/
├── model/      # データの形を定義
├── repository/ # データベース操作
├── usecase/    # ビジネスロジック
└── handler/    # HTTP処理
```

### 2. 依存関係の流れ

```go
// Handler → UseCase → Repository → Database
// 各層は下の層にのみ依存する

// Handler層
type BookHandler struct {
    bookUsecase usecase.BookUsecase  // UseCase層に依存
}

// UseCase層
type bookUsecase struct {
    bookRepo repository.BookRepository  // Repository層に依存
}

// Repository層
type bookRepository struct {
    db *database.DB  // Database層に依存
}
```

### 3. 実際のデータフロー例

```go
// 1. HTTPリクエストがHandlerに届く
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    // 2. JSONをGo構造体に変換
    var req model.CreateBookRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 3. UseCaseに処理を依頼
    book, err := h.bookUsecase.CreateBook(&req)
    
    // 4. 結果をJSONで返す
    h.sendSuccessResponse(w, http.StatusCreated, "成功", book)
}
```

## 🛠️ 実際に試してみよう

### 練習1: 構造体を作る

```go
package main

import "fmt"

// あなたの好きな本の情報を構造体で表現してみましょう
type MyBook struct {
    Title       string
    Author      string
    PageCount   int
    IsFinished  bool
}

func main() {
    // 実際の本の情報を入れてみてください
    book := MyBook{
        Title:      "ここにタイトル",
        Author:     "ここに著者名",
        PageCount:  300,
        IsFinished: false,
    }
    
    fmt.Printf("書籍: %s by %s\n", book.Title, book.Author)
    
    if book.IsFinished {
        fmt.Println("読了済み")
    } else {
        fmt.Println("未読または読書中")
    }
}
```

### 練習2: 簡単な関数を作る

```go
package main

import "fmt"

// 書籍の読書進捗を計算する関数
func calculateProgress(currentPage, totalPages int) float64 {
    if totalPages == 0 {
        return 0
    }
    return float64(currentPage) / float64(totalPages) * 100
}

func main() {
    progress := calculateProgress(150, 300)
    fmt.Printf("読書進捗: %.1f%%\n", progress)  // 50.0%
}
```

### 練習3: エラーハンドリング

```go
package main

import (
    "fmt"
    "errors"
)

func setRating(rating int) error {
    if rating < 1 || rating > 5 {
        return errors.New("評価は1-5の範囲で入力してください")
    }
    fmt.Printf("評価を%d点に設定しました\n", rating)
    return nil
}

func main() {
    // 正常なケース
    err := setRating(4)
    if err != nil {
        fmt.Println("エラー:", err)
    }
    
    // エラーケース
    err = setRating(10)
    if err != nil {
        fmt.Println("エラー:", err)
    }
}
```

## 🎓 次のステップ

1. **公式チュートリアル**: https://tour.golang.org/
2. **このプロジェクトのコードを読む**: コメントを参考にしながら実際のコードを理解
3. **小さな変更を加える**: メッセージやフィールドを変更してみる
4. **新機能を追加**: 「お気に入り」機能や「読書時間記録」機能を追加してみる

Go言語の世界へようこそ！🚀
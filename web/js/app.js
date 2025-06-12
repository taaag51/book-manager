// BookManager App - Main JavaScript

class BookManager {
    constructor() {
        this.apiBase = '/api/v1';
        this.currentPage = 1;
        this.pageSize = 20;
        this.currentFilters = {};
        this.editingBookId = null;
        this.currentReadingBookId = null;
        
        this.init();
    }

    async init() {
        this.bindEvents();
        await this.loadStatistics();
        await this.loadBooks();
        this.setTodayDate();
    }

    bindEvents() {
        // 書籍追加ボタン
        document.getElementById('addBookBtn').addEventListener('click', () => {
            this.showBookModal();
        });

        // モーダル関連
        document.getElementById('modalClose').addEventListener('click', () => {
            this.hideBookModal();
        });
        document.getElementById('cancelBtn').addEventListener('click', () => {
            this.hideBookModal();
        });
        document.getElementById('readingModalClose').addEventListener('click', () => {
            this.hideReadingModal();
        });
        document.getElementById('readingCancelBtn').addEventListener('click', () => {
            this.hideReadingModal();
        });

        // フォーム送信
        document.getElementById('bookForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.saveBook();
        });

        // 読書完了
        document.getElementById('finishReadingBtn').addEventListener('click', () => {
            this.finishReading();
        });

        // 検索・フィルター
        document.getElementById('searchInput').addEventListener('input', 
            this.debounce(() => this.applyFilters(), 300)
        );
        document.getElementById('statusFilter').addEventListener('change', () => this.applyFilters());
        document.getElementById('ratingFilter').addEventListener('change', () => this.applyFilters());

        // 表示切り替え
        document.querySelectorAll('.view-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                document.querySelectorAll('.view-btn').forEach(b => b.classList.remove('active'));
                e.target.classList.add('active');
                this.toggleView(e.target.dataset.view);
            });
        });

        // モーダル外クリックで閉じる
        document.getElementById('bookModal').addEventListener('click', (e) => {
            if (e.target.id === 'bookModal') this.hideBookModal();
        });
        document.getElementById('readingModal').addEventListener('click', (e) => {
            if (e.target.id === 'readingModal') this.hideReadingModal();
        });
    }

    // API呼び出し
    async apiCall(endpoint, options = {}) {
        try {
            this.showLoading();
            const response = await fetch(`${this.apiBase}${endpoint}`, {
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                throw new Error(data.error || 'APIエラーが発生しました');
            }
            
            return data;
        } catch (error) {
            this.showToast(error.message, 'error');
            throw error;
        } finally {
            this.hideLoading();
        }
    }

    // 統計情報読み込み
    async loadStatistics() {
        try {
            const response = await this.apiCall('/statistics');
            const stats = response.data;
            
            document.getElementById('totalBooks').textContent = stats.total_books || 0;
            document.getElementById('readingBooks').textContent = stats.reading_books || 0;
            document.getElementById('completedBooks').textContent = stats.completed_books || 0;
            document.getElementById('totalSpent').textContent = `¥${(stats.total_spent || 0).toLocaleString()}`;
        } catch (error) {
            console.error('統計情報の取得に失敗:', error);
        }
    }

    // 書籍一覧読み込み
    async loadBooks(page = 1) {
        try {
            const params = new URLSearchParams({
                page: page,
                limit: this.pageSize,
                ...this.currentFilters
            });
            
            const response = await this.apiCall(`/books?${params}`);
            this.renderBooks(response.data.books || []);
            this.renderPagination(response.data);
            this.currentPage = page;
        } catch (error) {
            console.error('書籍一覧の取得に失敗:', error);
            this.renderBooks([]);
        }
    }

    // 書籍一覧表示
    renderBooks(books) {
        const container = document.getElementById('booksContainer');
        
        if (books.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-book"></i>
                    <h3>書籍が見つかりません</h3>
                    <p>新しい書籍を追加してください</p>
                </div>
            `;
            return;
        }

        container.innerHTML = books.map(book => this.renderBookCard(book)).join('');
    }

    // 書籍カード表示
    renderBookCard(book) {
        const statusText = {
            'not_started': '未読',
            'reading': '読書中',
            'completed': '読了',
            'dropped': '中断'
        };

        const rating = book.rating ? '⭐'.repeat(book.rating) : '';
        const tags = book.tags ? book.tags.split(',').map(tag => 
            `<span class="tag">${tag.trim()}</span>`
        ).join('') : '';

        const purchaseDate = new Date(book.purchase_date).toLocaleDateString('ja-JP');
        
        return `
            <div class="book-card">
                <div class="book-header">
                    <div>
                        <h3 class="book-title">${this.escapeHtml(book.title)}</h3>
                        <p class="book-author">${this.escapeHtml(book.author)}</p>
                    </div>
                    <span class="book-status status-${book.status}">${statusText[book.status]}</span>
                </div>
                
                <div class="book-meta">
                    <div>購入日: ${purchaseDate}</div>
                    ${book.purchase_price ? `<div>価格: ¥${book.purchase_price.toLocaleString()}</div>` : ''}
                    ${book.publisher ? `<div>出版社: ${this.escapeHtml(book.publisher)}</div>` : ''}
                </div>
                
                ${rating ? `<div class="book-rating">${rating}</div>` : ''}
                ${tags ? `<div class="book-tags">${tags}</div>` : ''}
                ${book.notes ? `<div class="book-notes">${this.escapeHtml(book.notes)}</div>` : ''}
                
                <div class="book-actions">
                    <button class="btn btn-small btn-secondary" onclick="app.editBook(${book.id})">
                        <i class="fas fa-edit"></i> 編集
                    </button>
                    ${this.getActionButton(book)}
                    <button class="btn btn-small btn-danger" onclick="app.deleteBook(${book.id})">
                        <i class="fas fa-trash"></i> 削除
                    </button>
                </div>
            </div>
        `;
    }

    getActionButton(book) {
        switch (book.status) {
            case 'not_started':
                return `<button class="btn btn-small btn-success" onclick="app.startReading(${book.id})">
                    <i class="fas fa-play"></i> 読書開始
                </button>`;
            case 'reading':
                return `<button class="btn btn-small btn-primary" onclick="app.showFinishReadingModal(${book.id})">
                    <i class="fas fa-check"></i> 読書完了
                </button>`;
            default:
                return '';
        }
    }

    // ページネーション表示
    renderPagination(data) {
        const container = document.getElementById('pagination');
        const { page, total_pages: totalPages } = data;
        
        if (totalPages <= 1) {
            container.innerHTML = '';
            return;
        }

        let pagination = '';
        
        // 前へボタン
        pagination += `
            <button class="page-btn" ${page <= 1 ? 'disabled' : ''} 
                onclick="app.loadBooks(${page - 1})">
                <i class="fas fa-chevron-left"></i> 前へ
            </button>
        `;
        
        // ページ番号
        for (let i = Math.max(1, page - 2); i <= Math.min(totalPages, page + 2); i++) {
            pagination += `
                <button class="page-btn ${i === page ? 'active' : ''}" 
                    onclick="app.loadBooks(${i})">${i}</button>
            `;
        }
        
        // 次へボタン
        pagination += `
            <button class="page-btn" ${page >= totalPages ? 'disabled' : ''} 
                onclick="app.loadBooks(${page + 1})">
                次へ <i class="fas fa-chevron-right"></i>
            </button>
        `;
        
        container.innerHTML = pagination;
    }

    // フィルター適用
    applyFilters() {
        this.currentFilters = {};
        
        const search = document.getElementById('searchInput').value.trim();
        if (search) this.currentFilters.search = search;
        
        const status = document.getElementById('statusFilter').value;
        if (status) this.currentFilters.status = status;
        
        const rating = document.getElementById('ratingFilter').value;
        if (rating) this.currentFilters.rating = rating;
        
        this.currentPage = 1;
        this.loadBooks(1);
    }

    // 表示切り替え
    toggleView(view) {
        const container = document.getElementById('booksContainer');
        if (view === 'list') {
            container.classList.add('list-view');
        } else {
            container.classList.remove('list-view');
        }
    }

    // 書籍モーダル表示
    showBookModal(book = null) {
        this.editingBookId = book ? book.id : null;
        const modal = document.getElementById('bookModal');
        const form = document.getElementById('bookForm');
        const title = document.getElementById('modalTitle');
        
        title.textContent = book ? '書籍を編集' : '書籍を追加';
        
        if (book) {
            this.fillForm(book);
        } else {
            form.reset();
            this.setTodayDate();
        }
        
        modal.classList.add('show');
    }

    hideBookModal() {
        document.getElementById('bookModal').classList.remove('show');
        this.editingBookId = null;
    }

    // フォームに書籍データを設定
    fillForm(book) {
        document.getElementById('title').value = book.title || '';
        document.getElementById('author').value = book.author || '';
        document.getElementById('isbn').value = book.isbn || '';
        document.getElementById('publisher').value = book.publisher || '';
        document.getElementById('purchaseDate').value = book.purchase_date ? 
            book.purchase_date.split('T')[0] : '';
        document.getElementById('purchasePrice').value = book.purchase_price || '';
        document.getElementById('tags').value = book.tags || '';
        document.getElementById('notes').value = book.notes || '';
    }

    // 今日の日付を設定
    setTodayDate() {
        const today = new Date().toISOString().split('T')[0];
        document.getElementById('purchaseDate').value = today;
    }

    // 書籍保存
    async saveBook() {
        try {
            const formData = new FormData(document.getElementById('bookForm'));
            const data = {};
            
            for (let [key, value] of formData.entries()) {
                if (value.trim()) {
                    if (key === 'purchase_price') {
                        data[key] = parseInt(value);
                    } else if (key === 'purchase_date') {
                        data[key] = new Date(value).toISOString();
                    } else {
                        data[key] = value;
                    }
                }
            }
            
            if (this.editingBookId) {
                await this.apiCall(`/books/${this.editingBookId}`, {
                    method: 'PUT',
                    body: JSON.stringify(data)
                });
                this.showToast('書籍が更新されました', 'success');
            } else {
                await this.apiCall('/books', {
                    method: 'POST',
                    body: JSON.stringify(data)
                });
                this.showToast('書籍が追加されました', 'success');
            }
            
            this.hideBookModal();
            await this.loadBooks(this.currentPage);
            await this.loadStatistics();
        } catch (error) {
            console.error('書籍の保存に失敗:', error);
        }
    }

    // 書籍編集
    async editBook(id) {
        try {
            const response = await this.apiCall(`/books/${id}`);
            this.showBookModal(response.data);
        } catch (error) {
            console.error('書籍の取得に失敗:', error);
        }
    }

    // 書籍削除
    async deleteBook(id) {
        if (!confirm('この書籍を削除しますか？')) return;
        
        try {
            await this.apiCall(`/books/${id}`, { method: 'DELETE' });
            this.showToast('書籍が削除されました', 'success');
            await this.loadBooks(this.currentPage);
            await this.loadStatistics();
        } catch (error) {
            console.error('書籍の削除に失敗:', error);
        }
    }

    // 読書開始
    async startReading(id) {
        try {
            await this.apiCall(`/books/${id}/start-reading`, { method: 'POST' });
            this.showToast('読書を開始しました', 'success');
            await this.loadBooks(this.currentPage);
            await this.loadStatistics();
        } catch (error) {
            console.error('読書開始に失敗:', error);
        }
    }

    // 読書完了モーダル表示
    showFinishReadingModal(id) {
        this.currentReadingBookId = id;
        document.getElementById('readingModal').classList.add('show');
        // 評価をクリア
        document.querySelectorAll('input[name="rating"]').forEach(radio => {
            radio.checked = false;
        });
    }

    hideReadingModal() {
        document.getElementById('readingModal').classList.remove('show');
        this.currentReadingBookId = null;
    }

    // 読書完了
    async finishReading() {
        try {
            const rating = document.querySelector('input[name="rating"]:checked')?.value;
            const data = rating ? { rating: parseInt(rating) } : {};
            
            await this.apiCall(`/books/${this.currentReadingBookId}/finish-reading`, {
                method: 'POST',
                body: JSON.stringify(data)
            });
            
            this.showToast('読書を完了しました', 'success');
            this.hideReadingModal();
            await this.loadBooks(this.currentPage);
            await this.loadStatistics();
        } catch (error) {
            console.error('読書完了に失敗:', error);
        }
    }

    // ユーティリティ関数
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    showLoading() {
        document.getElementById('loadingOverlay').classList.add('show');
    }

    hideLoading() {
        document.getElementById('loadingOverlay').classList.remove('show');
    }

    showToast(message, type = 'success') {
        // 既存のトーストを削除
        const existingToast = document.querySelector('.toast');
        if (existingToast) {
            existingToast.remove();
        }

        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.innerHTML = `
            <i class="fas fa-${type === 'success' ? 'check' : 'exclamation'}-circle"></i>
            ${message}
        `;
        
        document.body.appendChild(toast);
        
        setTimeout(() => {
            toast.remove();
        }, 3000);
    }
}

// アプリケーション初期化
const app = new BookManager();
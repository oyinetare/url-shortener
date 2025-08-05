package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Compile-time check that Repository implements RepositoryInterface
var _ RepositoryInterface = (*Repository)(nil)

type Repository struct {
	db *sql.DB
}

// Connect creates a new repository connection
func Connect(host, database, user, password string, port int) (*Repository, error) {
	// Data Source Name - "connection string" to describe exactly how to reach and authenticate in mysql db
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// https://go.dev/doc/database/cancel-operations
	// https://go.dev/blog/contex

	// use context.Context pattern/package to easily pass request-scoped values, manage the lifecycle of operations,
	// coordinate cancellation signals/cancellation propagation, graceful shutdown, timeout handling and
	// deadlines across API boundaries to all the goroutines involved in handling a reques

	// so that when a request is canceled or times out, all the goroutines working on that request should
	// exit quickly so the system can reclaim any resources they are using.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Repository{db: db}, nil
}

// SaveUrls saves a new URL mapping
func (r *Repository) SaveUrls(ctx context.Context, shortUrl, longUrl string) error {
	// using prepared statements - https://go.dev/doc/database/prepared-statements
	query := `
		INSERT INTO urls (shortUrl, longUrl, createdAt, clicks)
		VALUES (?, ?, NOW(), 0)
	`

	_, err := r.db.ExecContext(ctx, query, shortUrl, longUrl)

	if err != nil {
		// Check for duplicate key error
		// 1062 is MySQL's error code for duplicate entry violations on unique constraints or primary keys
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == MySQLDuplicateEntry {
			return ErrDuplicateShortCode
		}
		return fmt.Errorf("failed to save URL: %w", err)
	}

	return nil
}

// GetShortURLFromLong retrieves a short URL by its long URL
func (r *Repository) GetShortURLFromLong(ctx context.Context, longUrl string) (*URLs, error) {
	// using prepared statements - https://go.dev/doc/database/prepared-statements
	// to prevent SQL Injection, improve performance & Type Safety
	var urls URLs
	query := `
		SELECT id, shortUrl, longUrl
		FROM urls
		WHERE longUrl = ?
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, longUrl).Scan(&urls.ID, &urls.ShortURL, &urls.LongURL)

	if err == sql.ErrNoRows {
		return nil, ErrURLNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get short URL: %w", err)
	}

	return &urls, nil
}

// GetLongURLFromShort retrieves a long URL by its short URL
func (r *Repository) GetLongURLFromShort(ctx context.Context, shortUrl string) (*URLs, error) {
	// using prepared statements - https://go.dev/doc/database/prepared-statements
	// to prevent SQL Injection, improve performance & Type Safety
	var urls URLs
	query := `
		SELECT id, shortUrl, longUrl
		FROM urls
		WHERE shortUrl = ?
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, shortUrl).Scan(&urls.ID, &urls.ShortURL, &urls.LongURL)

	if err == sql.ErrNoRows {
		return nil, ErrURLNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get short URL: %w", err)
	}

	return &urls, nil
}

// IncrementClicks increments the click count for a short URL
func (r *Repository) IncrementClicks(ctx context.Context, shortUrl string) error {
	// using prepared statements - https://go.dev/doc/database/prepared-statements
	// to prevent SQL Injection, improve performance & Type Safety
	query := `UPDATE urls SET clicks = clicks + 1 WHERE shortUrl = ?`
	result, err := r.db.ExecContext(ctx, query, shortUrl)

	if err != nil {
		return fmt.Errorf("failed to increment clicks: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrURLNotFound
	}

	return nil
}

// Disconnect closes the database connection
func (r *Repository) Disconnect() error {
	return r.db.Close()
}

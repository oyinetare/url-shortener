package repository

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

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

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

func (r *Repository) SaveUrls(shortUrl, longUrl string) error {
	query := `
		INSERT INTO urls (shortUrl, longUrl, createdAt, clicks)
		VALUES (?, ?, NOW(), 0)
	`

	_, err := r.db.Exec(query, shortUrl, longUrl)
	if err != nil {
		// Check for duplicate key error
		// if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
		// 	return errors.New("short code already exists")
		// }
		return fmt.Errorf("failed to save URL: %w", err)
	}

	return nil
}

func (r *Repository) GetShortURLFromLong(longUrl string) (*URLs, error) {
	var urls URLs
	query := `
		SELECT shortUrl, longUrl
		FROM urls
		WHERE longUrl = ?
		LIMIT 1
	`

	err := r.db.QueryRow(query, longUrl).Scan(&urls.ShortURL, &urls.LongURL)

	if err == sql.ErrNoRows {
		return nil, errors.New("url not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get short URL: %w", err)
	}

	return &urls, nil
}

func (r *Repository) GetLongURLFromShort(shortUrl string) (*URLs, error) {
	var urls URLs
	query := `
		SELECT shortUrl, longUrl
		FROM urls
		WHERE shortUrl = ?
		LIMIT 1
	`

	err := r.db.QueryRow(query, shortUrl).Scan(&urls.ShortURL, &urls.LongURL)

	if err == sql.ErrNoRows {
		return nil, errors.New("url not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get short URL: %w", err)
	}

	return &urls, nil
}

func (r *Repository) IncrementClicks(shortUrl string) error {
	query := `UPDATE urls SET clicks = clicks + 1 WHERE shortUrl = ?`
	_, err := r.db.Exec(query, shortUrl)
	return err
}

// Disconnect closes the database connection
func (r *Repository) Disconnect() error {
	return r.db.Close()
}

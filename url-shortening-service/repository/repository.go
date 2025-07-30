package repository

import (
	"database/sql"
	"fmt"
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

// Disconnect closes the database connection
func (r *Repository) Disconnect() error {
	return r.db.Close()
}

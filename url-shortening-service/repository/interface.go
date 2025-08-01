package repository

import (
	"context"
	"errors"
)

// Custom static errors for better error handling
var (
	ErrURLNotFound        = errors.New("url not found")
	ErrDuplicateShortCode = errors.New("short code already exists")
	ErrInvalidURL         = errors.New("invalid url")
)

// Constants for MySQL Error Numbers
const (
	MySQLDuplicateEntry  = 1062
	MySQLDeadlock        = 1213
	MySQLLockWaitTimeout = 1205
)

// URLs represents a URL mapping
type URLs struct {
	ID        int64  `json:"id,omitempty"`
	ShortURL  string `json:"shortUrl"`
	LongURL   string `json:"longUrl"`
	CreatedAt string `json:"createdAt,omitempty"`
	Clicks    int    `json:"clicks,omitempty"`
}

// RepositoryInterface defines the contract for URL storage
// Also so Repository can be used in tests with dependency injection
type RepositoryInterface interface {
	SaveUrls(ctx context.Context, shortUrl, longUrl string) error
	GetShortURLFromLong(ctx context.Context, longUrl string) (*URLs, error)
	GetLongURLFromShort(ctx context.Context, shortUrl string) (*URLs, error)
	IncrementClicks(ctx context.Context, shortUrl string) error
	Disconnect() error
}

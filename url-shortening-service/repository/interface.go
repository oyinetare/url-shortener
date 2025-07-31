package repository

// URLs represents a URL mapping
type URLs struct {
	ShortURL string `json:"shortUrl"`
	LongURL  string `json:"longUrl"`
}

// RepositoryInterface defines the contract for URL storage
// Also so Repository can be used in tests with dependency injection
type RepositoryInterface interface {
	SaveUrls(shortUrl, longUrl string) error
	GetShortURLFromLong(longUrl string) (*URLs, error)
	GetLongURLFromShort(shortUrl string) (*URLs, error)
	IncrementClicks(shortUrl string) error
	Disconnect() error
}

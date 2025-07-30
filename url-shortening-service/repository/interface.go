package repository

type URLs struct {
}

// RepositoryInterface defines the contract for URL storage
// Also so Repository can be used in tests with dependency injection
type RepositoryInterface interface {
	Disconnect() error
}

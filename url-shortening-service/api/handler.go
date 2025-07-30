package api

import (
	"net/http"

	"github.com/oyinetare/url-shortener/config"
	"github.com/oyinetare/url-shortener/repository"
)

// UrlShortenerAPI handles HTTP requests for URL shortening
type UrlShortenerAPI struct {
	repo   repository.RepositoryInterface
	config *config.Config
}

func NewUrlShortenerAPI(repo repository.RepositoryInterface, cfg *config.Config) *UrlShortenerAPI {
	return &UrlShortenerAPI{
		repo:   repo,
		config: cfg,
	}
}

// ShortenHandler handles POST requests to shorten URLs
func (api *UrlShortenerAPI) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPost{
	// 	a
	// }
}

// RedirectHandler handles GET requests to redirect to long URLs
func (api *UrlShortenerAPI) RedirectHandler(w http.ResponseWriter, r *http.Request) {
}

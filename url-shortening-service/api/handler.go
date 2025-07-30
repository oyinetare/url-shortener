package api

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

// ShortenRequest represents a request to shorten a URL
type ShortenRequest struct {
	LongURL string `json:"longUrl"`
}

// ShortenResponse represents a response with shortened URL
type ShortenResponse struct {
	ShortURL string `json:"shortUrl"`
	LongURL  string `json:"longUrl"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// generateShortCode creates a random short code
func (api *UrlShortenerAPI) generateShortCode(longUrl string) (string, error) {
	hasher := md5.New()
	hasher.Write([]byte(longUrl))
	hash := hasher.Sum(nil)

	// convert to base 64 helps so same number between
	// its different number representation systems
	// also makes URL-safe
	encoded := base64.URLEncoding.EncodeToString(hash)

	// Take first 7 characters and remove any special chars
	shortCode := strings.ReplaceAll(encoded[:api.config.ShortCodeLength], "/", "_")
	shortCode = strings.ReplaceAll(shortCode, "+", "-")
	shortCode = strings.ReplaceAll(shortCode, "=", "")

	return shortCode, nil
}

// shorten creates a short URL for the given long URL
func (api *UrlShortenerAPI) shorten(longUrl string) (string, error) {
	// if in hashmap return
	existing, err := api.repo.GetShortURLFromLong(longUrl)
	if err != nil && existing != nil {
		fullURL := fmt.Sprintf("%s/%s", api.config.BaseURL, existing.ShortURL)
		return fullURL, nil
	}

	// else generate shortCode with collision detection
	var shortCode string
	shortCode, err = api.generateShortCode(longUrl)
	if err != nil {
		return "", fmt.Errorf("failed to generate short code: %w", err)
	}

	// save to db
	err = api.repo.SaveUrls(shortCode, longUrl)
	if err != nil {
		return "", fmt.Errorf("failed to create unique short code")
	}

	fullURL := fmt.Sprintf("%s/%s", api.config.BaseURL, shortCode)
	return fullURL, nil
}

// ShortenHandler handles POST requests to shorten URLs
func (api *UrlShortenerAPI) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	shortURL, err := api.shorten(req.LongURL)
	if err != nil {
		log.Printf("Error shortening URL: %v", err)
		api.respondWithError(w, http.StatusInternalServerError, "Failed to shorten URL")
		return
	}

	resp := ShortenResponse{
		ShortURL: shortURL,
		LongURL:  req.LongURL,
	}

	api.respondWithJSON(w, http.StatusCreated, resp)
}

// RedirectHandler handles GET requests to redirect to long URLs
func (api *UrlShortenerAPI) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// get shortCode from path
	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == "" {
		api.respondWithError(w, http.StatusBadRequest, "Short code required")
		return
	}

	// find longUrl
	urlData, err := api.repo.GetLongURLFromShort(shortCode)
	if err != nil {
		// use fallback if not found
		http.NotFound(w, r)
		return
	}

	// defer func() {
	// 	if err := repo.Disconnect(); err != nil {
	// 		log.Printf("Error disconnecting from database: %v", err)
	// 	}
	// }()

	err = api.repo.IncrementClicks(shortCode)
	if err != nil {
		api.respondWithError(w, http.StatusBadRequest, "Problem incrementing clicks")
	}

	http.Redirect(w, r, urlData.LongURL, http.StatusFound)
}

// helper fucntion to sends a JSON response
func (api *UrlShortenerAPI) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// respondWithError sends an error response
func (api *UrlShortenerAPI) respondWithError(w http.ResponseWriter, code int, message string) {
	api.respondWithJSON(w, code, ErrorResponse{Error: message})
}

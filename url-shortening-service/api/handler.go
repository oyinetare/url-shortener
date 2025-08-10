package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/oyinetare/url-shortener/cache"
	"github.com/oyinetare/url-shortener/idgenerator"
	"github.com/oyinetare/url-shortener/repository"
)

// UrlShortenerAPI handles HTTP requests for URL shortening
type UrlShortenerAPI struct {
	repo        repository.RepositoryInterface
	baseURL     string
	idgenerator idgenerator.IDGeneratorInterface
	cache       cache.CacheInterface
}

func NewUrlShortenerAPI(repo repository.RepositoryInterface, baseURL string, idGen idgenerator.IDGeneratorInterface, cache cache.CacheInterface) *UrlShortenerAPI {
	return &UrlShortenerAPI{
		repo:        repo,
		baseURL:     baseURL,
		idgenerator: idGen,
		cache:       cache,
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

// shorten creates a short URL for the given long URL
func (api *UrlShortenerAPI) shorten(ctx context.Context, longUrl string) (string, error) {
	// Validate URL
	parsedURL, err := url.Parse(longUrl)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", repository.ErrInvalidURL
	}

	// Check if URL already exists
	existing, err := api.repo.GetShortURLFromLong(ctx, longUrl)
	if err == nil && existing != nil {
		fullURL := fmt.Sprintf("%s/%s", api.baseURL, existing.ShortURL)
		return fullURL, nil
	}

	// else generate shortCode with collision detection
	var shortCode string
	maxAttempts := 5

	for i := 0; i < maxAttempts; i++ {

		shortCode, err = api.idgenerator.GenerateShortCode(longUrl)
		if err != nil {
			return "", fmt.Errorf("failed to generate short code: %w", err)

		}

		// try save to db
		err = api.repo.SaveUrls(ctx, shortCode, longUrl)

		if err == nil {
			// Cache the new mapping
			api.cache.Set(shortCode, longUrl)
			break
		}

		if err != repository.ErrDuplicateShortCode {
			return "", err
		}
		// If duplicate, try again
	}

	if err != nil {
		return "", fmt.Errorf("failed to create unique short code after %d attempts", maxAttempts)
	}

	fullURL := fmt.Sprintf("%s/%s", api.baseURL, shortCode)
	return fullURL, nil
}

// ShortenHandler handles POST requests to shorten URLs
func (api *UrlShortenerAPI) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// https://go.dev/doc/database/cancel-operations
	// context.Context pattern for managing the lifecycle of operations and coordinated cancellation and timeout handling
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	shortURL, err := api.shorten(ctx, req.LongURL)
	if err != nil {
		log.Printf("Error shortening URL: %v", err)
		switch err {
		case repository.ErrInvalidURL:
			api.respondWithError(w, http.StatusBadRequest, "Invalid URL provided")
		default:
			api.respondWithError(w, http.StatusInternalServerError, "Failed to shorten URL")
		}
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

	// https://go.dev/doc/database/cancel-operations
	// context.Context pattern for managing the lifecycle of operations and coordinated cancellation and timeout handling
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// get shortCode from path
	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == "" {
		api.respondWithError(w, http.StatusBadRequest, "Short code required")
		return
	}

	// Check cache first
	if longURL, found := api.cache.Get(shortCode); found {
		log.Printf("Cache hit for short code: %s", shortCode)

		// Still increment clicks asynchronously
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := api.repo.IncrementClicks(ctx, shortCode); err != nil {
				log.Printf("Failed to increment clicks for %s: %v", shortCode, err)
			}
		}()

		http.Redirect(w, r, longURL, http.StatusFound)
		return
	}

	// Cache miss - fetch from database
	log.Printf("Cache miss for short code: %s", shortCode)

	// find longUrl
	urlData, err := api.repo.GetLongURLFromShort(ctx, shortCode)
	if err != nil {
		switch err {
		case repository.ErrURLNotFound:
			// use fallback if not found
			http.NotFound(w, r)
		case repository.ErrInvalidURL:
			api.respondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("Unexpected error: %v", err)
			api.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve URL")
		}
		return
	}

	// Update cache
	api.cache.Set(shortCode, urlData.LongURL)

	// increment click count with goroutine - best not to wait for it
	go func() {
		// https://go.dev/doc/database/cancel-operations
		// context.Context pattern for managing the lifecycle of operations and coordinated cancellation and timeout handling
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := api.repo.IncrementClicks(ctx, shortCode); err != nil {
			// fire and forget with Logging
			// not good practice to ever write to http.ResponseWriter from a goroutine after the handler returns
			log.Printf("Failed to increment clicks for %s: %v", shortCode, err)
		}
	}()

	// redirect to long URL
	http.Redirect(w, r, urlData.LongURL, http.StatusFound)
}

// respondWithJSON is helper fucntion to send a JSON response
func (api *UrlShortenerAPI) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// respondWithError is helper fucntion to send an error response
func (api *UrlShortenerAPI) respondWithError(w http.ResponseWriter, code int, message string) {
	api.respondWithJSON(w, code, ErrorResponse{Error: message})
}

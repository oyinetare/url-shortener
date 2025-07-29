package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type UrlShortener struct {
	port        string
	pathsToUrls map[string]string
}

// JSON request for shortening
type ShortenRequest struct {
	LongURL string `json:"longUrl"`
}

// JSON response for shortening
type ShortenResponse struct {
	ShortURL string `json:"shortUrl"`
	LongURL  string `json:"longUrl"`
}

func NewUrlShortener(port string) *UrlShortener {
	return &UrlShortener{
		port:        port,
		pathsToUrls: make(map[string]string),
	}
}

// generate short code using md5 as hashing function and encoding with base 64 conversion
func (u *UrlShortener) generateShortCode(longUrl string) string {

	hasher := md5.New()
	hasher.Write([]byte(longUrl))
	hash := hasher.Sum(nil)

	// convert to base 64 helps so same number between
	// its different number representation systems
	encoded := base64.URLEncoding.EncodeToString(hash)

	// Take first 7 characters and remove any special chars
	shortCode := strings.ReplaceAll(encoded[:7], "/", "_")
	shortCode = strings.ReplaceAll(shortCode, "+", "-")

	return shortCode
}

func (u *UrlShortener) Shorten(longUrl string) string {
	// if in hashmap return
	if shortUrl, exists := u.pathsToUrls[longUrl]; exists {
		return u.port + "/" + shortUrl
	}

	// else generate shortCode and save in hashmap
	shortUrl := u.generateShortCode(longUrl)

	u.pathsToUrls[shortUrl] = longUrl

	return u.port + "/" + shortUrl
}

func (u *UrlShortener) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	// get longurl
	var req ShortenRequest
	json.NewDecoder(r.Body).Decode(&req)

	shortUrl := u.Shorten(req.LongURL)

	resp := ShortenResponse{
		ShortURL: shortUrl,
		LongURL:  req.LongURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (u *UrlShortener) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// get shortUrl
	shortUrl := strings.TrimPrefix(r.URL.Path, "/")

	// if short code is there redirect to it
	if longUrl, exists := u.pathsToUrls[shortUrl]; exists {
		http.Redirect(w, r, longUrl, http.StatusFound)
		return
	}

	// use fallback if not found
	http.NotFound(w, r)
}

func main() {
	urlShortener := NewUrlShortener("http://localhost:8080")

	// register handlers
	http.HandleFunc("/shorten", urlShortener.ShortenHandler)
	http.HandleFunc("/", urlShortener.RedirectHandler)

	port := ":8080"
	fmt.Println("URL Shortener started on", port)
	fmt.Println("\nAPI Endpoints:")
	fmt.Println("POST /shorten - Shorten a URL")
	fmt.Println("GET  /{shortUrl}         - Redirect to long URL")
	fmt.Println("\nExample curl command:")
	fmt.Printf("curl -X POST http://localhost%s/shorten \\\n", port)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"longUrl":"https://www.google.com"}'`)

	// start server
	http.ListenAndServe(port, nil)
}

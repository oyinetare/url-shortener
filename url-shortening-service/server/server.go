package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/oyinetare/url-shortener/api"
	"github.com/oyinetare/url-shortener/config"
	"github.com/oyinetare/url-shortener/repository"
)

type Server struct {
	repo   repository.RepositoryInterface
	router *mux.Router
	config *config.Config
}

func New(repo repository.RepositoryInterface, config *config.Config) *Server {
	return &Server{
		repo:   repo,
		config: config,
	}
}

func (s *Server) Start() error {
	s.router = mux.NewRouter()

	// add logging middleware
	s.router.Use(loggingMiddleware)

	// initialise API handler and register routes
	shortenerAPI := api.NewUrlShortenerAPI(s.repo, s.config)

	s.router.HandleFunc("/shorten", shortenerAPI.ShortenHandler).Methods("POST")
	s.router.HandleFunc("/{shortCode}", shortenerAPI.RedirectHandler).Methods("GET")

	fmt.Printf("\n🚀 URL Shortener started on port %d\n", s.config.Port)
	fmt.Printf("📍 Base URL: %s\n", s.config.BaseURL)
	fmt.Printf("🔤 Short code length: %d\n\n", s.config.ShortCodeLength)

	fmt.Println("API Endpoints:")
	fmt.Println("POST /shorten      - Shorten a URL")
	fmt.Println("GET  /{shortCode}  - Redirect to long URL")
	fmt.Println("\nExample curl command:")
	fmt.Printf("curl -X POST %s/shorten \\\n", s.config.BaseURL)
	fmt.Println(`  -H "Content-Type: application/json" \`)
	fmt.Println(`  -d '{"longUrl":"https://www.example.com"}'`)

	addr := fmt.Sprintf(":%d", s.config.Port)
	log.Printf("Server starting on %s", addr)

	return http.ListenAndServe(addr, s.router)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

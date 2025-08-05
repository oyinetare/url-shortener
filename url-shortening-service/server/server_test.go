package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/oyinetare/url-shortener/config"
	"github.com/oyinetare/url-shortener/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveUrls(ctx context.Context, shortUrl, longUrl string) error {
	args := m.Called()
	if args.Get(0) == nil {
		return args.Error(1)
	}

	return args.Error(1)
}
func (m *MockRepository) GetShortURLFromLong(ctx context.Context, longUrl string) (*repository.URLs, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*repository.URLs), args.Error(1)
}
func (m *MockRepository) GetLongURLFromShort(ctx context.Context, shortUrl string) (*repository.URLs, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*repository.URLs), args.Error(1)
}
func (m *MockRepository) IncrementClicks(ctx context.Context, shortUrl string) error {
	args := m.Called()
	if args.Get(0) == nil {
		return args.Error(1)
	}

	return args.Error(1)
}

func (m *MockRepository) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func TestNew(t *testing.T) {
	mockRepo := new(MockRepository)

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8080
	}
	baseUrl := os.Getenv("BASE_URL")

	cfg := &config.Config{
		Port:    port,
		BaseURL: baseUrl,
	}

	srv := New(mockRepo, cfg)

	assert.NotNil(t, srv)
	assert.Equal(t, mockRepo, srv.repo)
	assert.Equal(t, cfg, srv.config)
}

func TestRoutes(t *testing.T) {
	mockRepo := new(MockRepository)

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8080
	}
	baseUrl := os.Getenv("BASE_URL")

	cfg := &config.Config{
		Port:    port,
		BaseURL: baseUrl,
	}

	srv := New(mockRepo, cfg)

	// start server in goroutine
	go func() {
		srv.Start()
	}()

	// give server time to sleep
	time.Sleep(100 * time.Millisecond)

	// test registered routes
	assert.NotNil(t, srv.router)
}

func TestLoggingMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// test that middleware doesn't break the request
	loggingMiddleware(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

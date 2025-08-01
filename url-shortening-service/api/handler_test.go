package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/oyinetare/url-shortener/config"
	"github.com/oyinetare/url-shortener/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveUrls(ctx context.Context, shortUrl, longUrl string) error {
	args := m.Called(ctx, shortUrl, longUrl)
	return args.Error(0)
}

func (m *MockRepository) GetShortURLFromLong(ctx context.Context, longUrl string) (*repository.URLs, error) {
	args := m.Called(ctx, longUrl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.URLs), args.Error(1)
}

func (m *MockRepository) GetLongURLFromShort(ctx context.Context, shortUrl string) (*repository.URLs, error) {
	args := m.Called(ctx, shortUrl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.URLs), args.Error(1)
}

func (m *MockRepository) IncrementClicks(ctx context.Context, shortUrl string) error {
	args := m.Called(ctx, shortUrl)
	return args.Error(0)
}

func (m *MockRepository) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*MockRepository)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:        "successful shortening",
			requestBody: `{"longUrl":"https://example.com"}`,
			mockSetup: func(m *MockRepository) {
				m.On("GetShortURLFromLong", mock.Anything, "https://example.com").
					Return(nil, repository.ErrURLNotFound)
				m.On("SaveUrls", mock.Anything, mock.Anything, "https://example.com").
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"longUrl": "https://example.com",
			},
		},
		{
			name:        "URL already exists",
			requestBody: `{"longUrl":"https://example.com"}`,
			mockSetup: func(m *MockRepository) {
				m.On("GetShortURLFromLong", mock.Anything, "https://example.com").
					Return(&repository.URLs{
						ShortURL: "existing123",
						LongURL:  "https://example.com",
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"shortUrl": "http://localhost:8080/existing123",
				"longUrl":  "https://example.com",
				// "shortCode": "existing123",
			},
		},
		// {
		// 	name:           "invalid URL",
		// 	requestBody:    `{"longUrl":"not-a-url"}`,
		// 	mockSetup:      func(m *MockRepository) {},
		// 	expectedStatus: http.StatusBadRequest,
		// 	expectedBody: map[string]interface{}{
		// 		"error": "Invalid URL provided",
		// 	},
		// },
		{
			name:           "invalid JSON",
			requestBody:    `{invalid json}`,
			mockSetup:      func(m *MockRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request body",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)

			cfg := &config.Config{
				Port:            8080,
				BaseURL:         "http://localhost:8080",
				ShortCodeLength: 7,
			}
			api := NewUrlShortenerAPI(mockRepo, cfg, cfg.ShortCodeLength)

			// Create request
			req := httptest.NewRequest("POST", "/shorten", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			api.ShortenHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Check expected fields
			for key, value := range tt.expectedBody {
				if key == "shortUrl" || key == "shortCode" {
					// For generated values, just check they exist
					if tt.expectedStatus == http.StatusCreated && tt.name == "successful shortening" {
						assert.NotEmpty(t, response[key])
						continue
					}
				}
				assert.Equal(t, value, response[key])
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	tests := []struct {
		name           string
		shortCode      string
		mockSetup      func(*MockRepository)
		expectedStatus int
		expectedHeader string
	}{
		{
			name:      "successful redirect",
			shortCode: "abc123",
			mockSetup: func(m *MockRepository) {
				m.On("GetLongURLFromShort", mock.Anything, "abc123").
					Return(&repository.URLs{
						ShortURL: "abc123",
						LongURL:  "https://example.com",
					}, nil)
				m.On("IncrementClicks", mock.Anything, "abc123").
					Return(nil).Maybe()
			},
			expectedStatus: http.StatusFound,
			expectedHeader: "https://example.com",
		},
		{
			name:      "URL not found",
			shortCode: "notfound",
			mockSetup: func(m *MockRepository) {
				m.On("GetLongURLFromShort", mock.Anything, "notfound").
					Return(nil, repository.ErrURLNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)

			cfg := &config.Config{
				BaseURL:         "http://localhost:8080",
				ShortCodeLength: 7,
			}
			api := NewUrlShortenerAPI(mockRepo, cfg, cfg.ShortCodeLength)

			// Create request with gorilla mux
			req := httptest.NewRequest("GET", "/"+tt.shortCode, nil)
			w := httptest.NewRecorder()

			// Set up router to capture path parameter
			router := mux.NewRouter()
			router.HandleFunc("/{shortCode}", api.RedirectHandler)
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, w.Header().Get("Location"))
			}

			// Allow time for async increment
			time.Sleep(10 * time.Millisecond)
			mockRepo.AssertExpectations(t)
		})
	}
}

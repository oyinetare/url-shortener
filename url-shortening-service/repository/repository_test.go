package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// mock ping
	mock.ExpectPing()

	// need to test the connection logic without actually connecting
	// bur would require refactoring the Connect function to accept a db interface
	// For now, we'll test the error cases
	t.Run("returns error on connection failure", func(t *testing.T) {
		_, err := Connect("", "", "", "", 0)
		assert.Error(t, err)
	})
}

func TestDisconnect(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	// defer db.Close() DONT NEED THIS?

	repo := &Repository{db: db}

	mock.ExpectClose()

	err = repo.Disconnect()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_SaveUrls(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: db}
	ctx := context.Background()

	tests := []struct {
		name      string
		shortURL  string
		longURL   string
		mockSetup func()
		wantErr   error
	}{
		{
			name:     "successful save",
			shortURL: "abc123",
			longURL:  "https://example.com",
			mockSetup: func() {
				mock.ExpectExec("INSERT INTO urls").
					WithArgs("abc123", "https://example.com").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: nil,
		},
		{
			name:     "duplicate short code",
			shortURL: "abc123",
			longURL:  "https://example.com",
			mockSetup: func() {
				mock.ExpectExec("INSERT INTO urls").
					WithArgs("abc123", "https://example.com").
					WillReturnError(&mysql.MySQLError{Number: 1062})
			},
			wantErr: ErrDuplicateShortCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := repo.SaveUrls(ctx, tt.shortURL, tt.longURL)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestIncrementClicks(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: db}
	ctx := context.Background()

	tests := []struct {
		name      string
		shortURL  string
		mockSetup func()
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "successful increment",
			shortURL: "abc123",
			mockSetup: func() {
				mock.ExpectExec("UPDATE urls SET clicks = clicks \\+ 1 WHERE shortUrl = \\?").
					WithArgs("abc123").
					WillReturnResult(sqlmock.NewResult(0, 1)) // 0 = lastInsertId, 1 = rowsAffected
			},
			wantErr: false,
		},
		{
			name:     "URL not found (no rows affected)",
			shortURL: "notfound",
			mockSetup: func() {
				mock.ExpectExec("UPDATE urls SET clicks = clicks \\+ 1 WHERE shortUrl = \\?").
					WithArgs("notfound").
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
			},
			wantErr: false, // Your current implementation doesn't check rowsAffected
		},
		{
			name:     "database error",
			shortURL: "abc123",
			mockSetup: func() {
				mock.ExpectExec("UPDATE urls SET clicks = clicks \\+ 1 WHERE shortUrl = \\?").
					WithArgs("abc123").
					WillReturnError(errors.New("database connection lost"))
			},
			wantErr: true,
			errMsg:  "database connection lost",
		}, {
			name:     "URL not found (no rows affected)",
			shortURL: "notfound",
			mockSetup: func() {
				mock.ExpectExec("UPDATE urls SET clicks = clicks \\+ 1 WHERE shortUrl = \\?").
					WithArgs("notfound").
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
			},
			wantErr: true,
			errMsg:  "url not found", // Now it returns ErrURLNotFound
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := repo.IncrementClicks(ctx, tt.shortURL)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)

				} else {
					assert.NoError(t, err)
				}

				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

func TestRepository_GetShortURLFromLong(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := &Repository{db: db}
	ctx := context.Background()

	tests := []struct {
		name      string
		longURL   string
		mockSetup func()
		want      *URLs
		wantErr   error
	}{
		{
			name:    "found URL",
			longURL: "https://example.com",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "shortUrl", "longUrl"}).
					AddRow(1, "abc123", "https://example.com")
				mock.ExpectQuery("SELECT id, shortUrl, longUrl FROM urls WHERE longUrl").
					WithArgs("https://example.com").
					WillReturnRows(rows)
			},
			want: &URLs{
				ID:       1,
				ShortURL: "abc123",
				LongURL:  "https://example.com",
			},
			wantErr: nil,
		},
		{
			name:    "URL not found",
			longURL: "https://notfound.com",
			mockSetup: func() {
				mock.ExpectQuery("SELECT id, shortUrl, longUrl FROM urls WHERE longUrl").
					WithArgs("https://notfound.com").
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: ErrURLNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			got, err := repo.GetShortURLFromLong(ctx, tt.longURL)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ShortURL, got.ShortURL)
				assert.Equal(t, tt.want.LongURL, got.LongURL)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

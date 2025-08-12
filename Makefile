.PHONY: db app test-db all clean

db:
	cd test-db && ./setup_db.sh

test-db:
	cd test-db && ./test_db.sh

app:
	cd url-shortening-service && go run main.go

build:
	cd url-shortening-service && go build -o ../url-shortener

all: db app

.PHONY: coverage
coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"
	@go tool cover -func=coverage.txt | grep total

clean:
	docker stop urls_db 2>/dev/null || true
	docker rm urls_db 2>/dev/null || true
	rm -f url-shortener

.PHONY: test test-verbose test-coverage test-coverage-html

test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func=coverage.txt

test-coverage-html:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration:
	go test -tags=integration ./...

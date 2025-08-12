# url-shortener (project not complete yet)

Toy URL shortener built with Go, featuring a layered architecture, MySQL database, in-memory caching, and comprehensive testing.

This project demonstrates modern microservice design patterns and best system design practices.

## 📋 Table of Contents

- [Project Overview](#-project-overview)
- [Project Structure](#-project-structure)
- [Getting Started](#-getting-started)
- [API Documentation](#-api-documentation)
- [Docker Commands](#-docker-commands)
- [Architecture & Design Patterns](#-architecture--design-patterns)
- [Implementation Steps](#-implementation-steps)
- [Features Implemented](#-features-implemented)
- [Testing](#-testing)
- [Future Roadmap](#-future-roadmap)

## 🎯 Project Overview

This URL shortener service allows users to:
- Convert long URLs into short, memorable links
- Redirect from short URLs to original destinations
- Track click statistics
- Cache frequently accessed URLs for improved performance

The project is built with a focus on scalability, maintainability, and production-readiness.

## 📁 Project Structure

```
├── test-database/      # MySQL Docker setup
├──── Dockerfile  
├──── setup_db.sh        # Database setup script
└──── setup.sql          # Database schema
├── url-shortening-service/
├──── api/                # HTTP handlers and API logic
│     ├── handler.go      # Request handlers
│     └── handler_test.go # Handler tests
├──── cache/              # In-memory caching implementation
│     ├── cache.go        # Cache logic with TTL
│     ├── cache_test.go   # Cache tests
│     └── interface.go    # Cache interface
├──── config/             # Configuration management
│     ├── config.go       # Config loader
│     └── config_test.go  # Config tests
├──── idgenerator/        # ID generation algorithms
│     ├── interface.go    # Generator interface
│     ├── md5Generator.go # MD5-based generator
│     └── snowflakeGenerator.go # Snowflake ID generator
├──── repository/         # Database access layer
│     ├── interface.go    # Repository interface
│     ├── repository.go   # MySQL implementation
│     └── repository_test.go # Repository tests
├──── server/             # Server setup and middleware
│      ├── server.go       # Server initialization
│      └── server_test.go  # Server tests
├──── .env.example        # Environment variables template
├──── Dockerfile          # Container configuration
└──── main.go           # Database schema
├── docker-compose.db.yml
├── docker-compose.yml  # Service orchestration
└── README.md           # Project documentation
```


## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- MySQL 8.0 (via Docker)

### Quick Start

1. **Clone the repository**
```bash
git clone https://github.com/oyinetare/url-shortener.git
cd url-shortener
```

2. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. **Start the database**
```bash
./setup_db.sh
# Or manually with Docker:
docker run --name urls_db -d \
  -e MYSQL_ROOT_PASSWORD=123 \
  -e MYSQL_DATABASE=urls \
  -e MYSQL_USER=url_shorten_service \
  -e MYSQL_PASSWORD=123 \
  -p 3306:3306 \
  mysql:8.0
```

4. **Run the application**
```bash
go run main.go -shortCode=7
```

5. **Or build and run with Docker**
```bash
# Build database
docker build -t test-db ./test-db

# Run database
docker run --name urls_db -p 3306:3306 test-db

# Build url shortening service
docker build -t url-shortener .

# Run url shortening service (linked to database)
docker run --name url-shortener -p 8080:8080 --link urls_db:db -e DATABASE_HOST=db url-shortener
```

### Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `PORT` | Server port | `8080` |
| `BASE_URL` | Base URL for short links | `http://localhost:8080` |
| `DATABASE_HOST` | MySQL host | `127.0.0.1` |
| `DATABASE_PORT` | MySQL port | `3306` |
| `DATABASE_NAME` | Database name | `urls` |
| `DATABASE_USER` | Database user | `url_shorten_service` |
| `DATABASE_PASSWORD` | Database password | `123` |
| `SHORT_CODE_LENGTH` | Length of short codes | `7` |
| `CACHE_TTL_MINUTES` | Cache TTL in minutes | `60` |

## 🐳 Docker Commands

### Docker Compose Commands

```bash
# Start all services
docker-compose up

# Start in background
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Rebuild and start
docker-compose up --build
```

### Useful Docker Commands

```bash
# List running containers
docker ps

# View container logs
docker logs <container_name>

# Execute command in container
docker exec -it <container_name> /bin/sh

# Remove stopped containers
docker container prune

# Remove unused images
docker image prune
```

## 🏗 Architecture & Design Patterns

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│  API Layer  │────▶│    Cache    │
└─────────────┘     │  (Handlers) │     │ (In-Memory) │
                    └──────┬──────┘     └──────┬──────┘
                           │                   │
                    ┌──────▼──────┐           │
                    │   Business   │           │
                    │    Logic     │◀──────────┘
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ Repository  │
                    │  Interface  │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │    MySQL     │
                    │   Database   │
                    └─────────────┘
```

### Layered Architecture

```
API Layer (HTTP Handlers)
    ↓
Business Logic Layer (Server)
    ↓
Data Access Layer (Repository)
    ↓
Database Layer (MySQL)
```

### Design Patterns

This project implements several design patterns:

1. **Repository Pattern** - Abstracts database access
2. **Dependency Injection** - Improves testability and flexibility
3. **Factory Pattern** - Creates configuration objects
4. **Layered Architecture** - Separates concerns
5. **Interface-Based Design** - Enables mocking and testing

## 📈 Implementation Steps

### Step 1: Basic Proof of Concept ✅
- Created initial HTTP server with basic routing
- Implemented simple URL shortening logic
- Set up project structure with Go modules
- Basic API endpoints for shortening and redirecting

### Step 2: Database & Architecture ✅
- **MySQL Integration with Docker**
  - Containerized MySQL database
  - Database schema design with proper indexes
  - Connection pooling and timeout handling

- **Layered Architecture Implementation**
  - **API Layer**: HTTP handlers with Gorilla Mux
  - **Business Logic Layer**: Server coordination
  - **Data Access Layer**: Repository pattern
  - **Database Layer**: MySQL with prepared statements

- **Dependency Injection**
  - Interface-based design for testability
  - Mock implementations for unit testing
  - Clean separation of concerns

### Step 3: Production Enhancements ✅
- **Context Pattern Implementation**
  - Request-scoped context for cancellation
  - Timeout handling across all database operations
  - Graceful shutdown support

- **Robust Error Handling**
  - Custom error types (ErrURLNotFound, ErrDuplicateShortCode, etc.)
  - Proper HTTP status code mapping
  - Detailed error logging

- **Configuration Management**
  - Environment-based configuration
  - Command-line flags for runtime options
  - .env file support with godotenv

- **Comprehensive Testing**
  - Unit tests with testify and go-sqlmock
  - Test coverage goals: ~85-100% per package
  - Race condition testing with `-race` flag

- **In-Memory Caching**
  - Thread-safe cache implementation with sync.RWMutex
  - Configurable TTL with automatic expiration
  - Background cleanup goroutine
  - Cache hit/miss logging for monitoring

## ✨ Features Implemented

- **URL Shortening**: MD5-based and Snowflake ID generation
- **Custom Short Codes**: Configurable length (default: 7 characters)
- **Click Tracking**: Asynchronous click count updates
- **Caching**: In-memory cache with TTL and automatic cleanup
- **Error Handling**: Comprehensive error types and HTTP status mapping
- **Configuration**: Environment variables and command-line flags
- **Testing**: Unit tests with mocks, ~85%+ coverage
- **Logging**: Request logging middleware
- **Database**: MySQL with prepared statements and connection pooling

## 🧪 Testing

### Run All Tests
```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Test Coverage Goals
- Config package: ~100% ✅
- Repository package: ~85% ✅
- API package: ~90% ✅
- Cache package: ~95% ✅

## 🗺 Future Roadmap

### Phase 1: Algorithm & Performance
- [ ] Add Bloom filter for quick existence checks
- [ ] Enhance URL validation
- [ ] Implement connection pooling optimizations

### Phase 2: Scalability
- [ ] Add Redis for distributed caching
- [ ] Implement database sharding
- [ ] Add read replicas
- [ ] Implement CQRS pattern
- [ ] Add message queue for analytics

### Phase 3: Infrastructure
- [ ] Complete Docker Compose setup
- [ ] Add Nginx reverse proxy
- [ ] Implement rate limiting
- [ ] Add API Gateway
- [ ] CDN integration for redirects

### Phase 4: Observability
- [ ] Add health check endpoints
- [ ] Implement structured logging (Zap/Logrus)
- [ ] Add Prometheus metrics
- [ ] Implement distributed tracing
- [ ] Create Grafana dashboards

### Phase 5: Production Readiness
- [ ] Add integration tests
- [ ] Implement CI/CD pipeline
- [ ] Implement disaster recovery

## 📝 License

MIT License

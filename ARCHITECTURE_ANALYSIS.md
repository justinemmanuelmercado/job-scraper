# Job Scraper Architecture Analysis & Recommendations

## Executive Summary

This document provides a comprehensive analysis of the Go-based job scraper project and proposes architectural improvements to enhance maintainability, testability, and scalability while following Go best practices.

## Current Architecture Analysis

### What the Original Author Intended

The project aims to create a job aggregation system that:
- Scrapes job postings from multiple sources (RSS feeds, Reddit, HackerNews)
- Deduplicates and stores jobs in PostgreSQL database
- Sends notifications via Discord when new jobs are found
- Provides markdown export functionality for job listings

### Current Architecture Overview

```
main.go (orchestrator)
├── pkg/rss_feed/     - RSS feed scraping
├── pkg/reddit/       - Reddit API integration  
├── pkg/hackernews/   - HackerNews scraping
├── pkg/store/        - Database operations
├── pkg/discord/      - Discord notifications
├── pkg/models/       - Data structures
└── pkg/errors.go     - Basic error handling
```

## Architectural Issues Identified

### 1. **Tight Coupling & Mixed Concerns**
- `main.go:55-104` - Orchestration mixed with business logic
- Database connections managed in multiple places
- Hard-coded configuration scattered throughout (e.g., `hackernews.go:34`, `discord.go:62`)
- No dependency injection or interface abstraction

### 2. **Inconsistent Error Handling**
- Custom error handler in `pkg/errors.go:8-16` but inconsistent usage
- Some functions return errors, others use fatal logging
- No structured error types or context propagation
- Mix of `log.Fatalf` and error returns (`main.go:60`, `main.go:79`)

### 3. **Database Layer Issues**
- No connection pooling (`store/db.go:9-17`)
- Manual query building with string formatting (`store/notice.go:23-48`)
- Mixed Go/TypeScript database management (Drizzle + Go)
- No transaction support or retry logic

### 4. **Limited Testability**
- Concrete types instead of interfaces
- Direct database connections in business logic
- Hard-coded external dependencies
- Only basic test in `pkg/reddit/reddit_test.go`

### 5. **Configuration Management**
- Environment variables accessed directly throughout code
- No validation or default values
- Mixed configuration sources (env vars + hardcoded values)

### 6. **Resource Management**
- No graceful shutdown handling
- Database connections not properly managed
- No context cancellation for long-running operations

## Recommended Architecture Improvements

### 1. **Clean Architecture Implementation**

```
cmd/
├── scraper/          - Main application entry
└── server/           - Optional web server

internal/
├── domain/           - Business entities and interfaces
│   ├── entities/     - Core business models
│   ├── repositories/ - Data access interfaces  
│   └── services/     - Business logic interfaces
├── usecases/         - Application business rules
├── adapters/         - External service implementations
│   ├── repositories/ - Database implementations
│   ├── scrapers/     - Scraper implementations
│   └── notifiers/    - Notification implementations
└── infrastructure/   - Framework and drivers
    ├── database/     - DB connection and migrations
    ├── config/       - Configuration management
    └── logging/      - Structured logging

pkg/                  - Public interfaces (if needed)
```

### 2. **Interface-Driven Design**

Define core interfaces:

```go
// domain/repositories/notice.go
type NoticeRepository interface {
    Create(ctx context.Context, notices []*domain.Notice) error
    GetLatest(ctx context.Context, limit int) ([]*domain.Notice, error)
    GetCount(ctx context.Context) (int, error)
}

// domain/services/scraper.go  
type Scraper interface {
    Name() string
    Scrape(ctx context.Context) ([]*domain.Notice, error)
}

// domain/services/notifier.go
type Notifier interface {
    SendJobs(ctx context.Context, jobs []*domain.Notice) error
    SendSummary(ctx context.Context, summary Summary) error
}
```

### 3. **Configuration Management**

```go
// internal/infrastructure/config/config.go
type Config struct {
    Database DatabaseConfig `validate:"required"`
    Reddit   RedditConfig   `validate:"required"`
    Discord  DiscordConfig  `validate:"required"`
    App      AppConfig
}

func Load() (*Config, error) {
    // Use viper or similar for configuration loading
    // Include validation and environment-specific overrides
}
```

### 4. **Dependency Injection**

```go
// internal/infrastructure/container.go
type Container struct {
    Config     *config.Config
    DB         *sql.DB
    Logger     *slog.Logger
    Scrapers   []services.Scraper
    Notifiers  []services.Notifier
    Repository repositories.NoticeRepository
}

func NewContainer(cfg *config.Config) (*Container, error) {
    // Initialize all dependencies with proper error handling
}
```

### 5. **Improved Error Handling**

```go
// pkg/errors/errors.go
type ErrorType string

const (
    ErrorTypeDatabase ErrorType = "database"
    ErrorTypeScraper  ErrorType = "scraper"
    ErrorTypeNetwork  ErrorType = "network"
)

type AppError struct {
    Type    ErrorType
    Message string
    Cause   error
    Context map[string]interface{}
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
}
```

### 6. **Use Cases Layer**

```go
// internal/usecases/scrape_jobs.go
type ScrapeJobsUseCase struct {
    repo      repositories.NoticeRepository
    scrapers  []services.Scraper
    notifiers []services.Notifier
    logger    *slog.Logger
}

func (uc *ScrapeJobsUseCase) Execute(ctx context.Context) error {
    // Business logic for scraping, deduplication, and notification
    // Proper error handling and logging throughout
}
```

### 7. **Database Improvements**

```go
// internal/adapters/repositories/notice_postgres.go
type PostgresNoticeRepository struct {
    db     *sql.DB
    logger *slog.Logger
}

func (r *PostgresNoticeRepository) Create(ctx context.Context, notices []*domain.Notice) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    // Use prepared statements and proper transaction handling
    // Include retry logic and connection pooling
}
```

## Implementation Benefits

### 1. **Maintainability**
- Clear separation of concerns
- Single responsibility principle adherence
- Easier to locate and modify specific functionality

### 2. **Testability**  
- Mockable interfaces for all external dependencies
- Isolated business logic testing
- Integration test capabilities

### 3. **Scalability**
- Pluggable scraper architecture
- Async processing capabilities
- Connection pooling and resource management

### 4. **Reliability**
- Proper error handling and recovery
- Graceful shutdown and resource cleanup
- Transaction support and data consistency

### 5. **Observability**
- Structured logging throughout
- Metrics and monitoring hooks
- Tracing support for debugging

## Migration Strategy

### Phase 1: Foundation
1. Introduce configuration management
2. Add structured logging
3. Create core domain interfaces

### Phase 2: Refactor Core Logic
1. Extract business logic to use cases
2. Implement interface-driven scrapers
3. Add proper error handling

### Phase 3: Infrastructure
1. Improve database layer with connection pooling
2. Add comprehensive testing
3. Implement dependency injection

### Phase 4: Enhancement
1. Add monitoring and metrics
2. Implement graceful shutdown
3. Add performance optimizations

## Conclusion

The current implementation shows good understanding of Go concurrency patterns but lacks architectural maturity. The proposed improvements will transform it into a maintainable, testable, and scalable system following Go best practices and clean architecture principles.

The refactoring preserves the original functionality while significantly improving code quality, error handling, and long-term maintainability.
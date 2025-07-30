# Job Scraper - Clean Architecture Implementation

A Go-based job scraper that aggregates job postings from multiple sources (RSS feeds, Reddit, HackerNews) and stores them in PostgreSQL with Discord notifications.

## Architecture Overview

This project has been refactored to follow Clean Architecture principles with proper separation of concerns, dependency injection, and comprehensive error handling.

### New Architecture Structure

```
├── cmd/                     # Application entry points (removed in favor of root main.go)
├── internal/                # Private application code
│   ├── domain/             # Business logic layer
│   │   ├── entities/       # Core business entities
│   │   ├── repositories/   # Data access interfaces
│   │   └── services/       # Business service interfaces
│   ├── usecases/          # Application business rules
│   ├── adapters/          # Interface adapters
│   │   ├── repositories/  # Database implementations
│   │   ├── scrapers/     # Scraper implementations
│   │   └── notifiers/    # Notification implementations
│   └── infrastructure/   # Frameworks and drivers
│       ├── config/       # Configuration management
│       ├── database/     # Database connection
│       ├── logging/      # Structured logging
│       ├── errors/       # Error handling
│       └── container/    # Dependency injection
├── pkg/                  # Legacy public code (kept for compatibility)
└── main.go              # Application entry point
```

## Key Improvements

### 1. Clean Architecture
- **Domain Layer**: Pure business logic with entities and interfaces
- **Use Cases**: Application-specific business rules
- **Adapters**: Interface implementations for external services
- **Infrastructure**: Framework-specific code and utilities

### 2. Dependency Injection
- Centralized container manages all dependencies
- Interface-driven design for easy testing
- Proper lifecycle management

### 3. Configuration Management  
- Environment-based configuration with validation
- Support for multiple deployment environments
- Centralized configuration loading

### 4. Structured Logging
- JSON-formatted logs with contextual information
- Component-based logging with structured fields
- Configurable log levels

### 5. Error Handling
- Custom error types with context
- Proper error wrapping and unwrapping
- Structured error information

### 6. Database Layer
- Connection pooling with configurable limits
- Transaction support
- Repository pattern implementation
- Prepared statements for performance

### 7. Concurrency & Performance
- Concurrent scraping of multiple sources
- Graceful shutdown handling
- Context-based cancellation
- Resource cleanup

## Getting Started

### Prerequisites
- Go 1.21+
- PostgreSQL
- Node.js and pnpm (for database management)

### Environment Setup

1. **Database Setup**:
   ```bash
   # Generate and apply migrations
   pnpm db:generate
   pnpm db:migrate
   ```

2. **Environment Variables**:
   Copy `.env.example` to `.env` and configure:
   ```env
   # Database
   DATABASE_URL=postgresql://user:password@localhost:5432/dbname
   
   # Reddit API
   REDDIT_ID=your_reddit_client_id
   REDDIT_SECRET=your_reddit_client_secret
   REDDIT_USERNAME=your_reddit_username
   REDDIT_PASSWORD=your_reddit_password
   
   # Discord (optional)
   JOBBYMCJOBFACE_DISCORD_TOKEN=your_discord_bot_token
   JOBBYMCJOBFACE_HIRING_CHANNEL_ID=channel_id
   DISCORD_ENABLED=true
   
   # Application
   LOG_LEVEL=info
   ENVIRONMENT=development
   ```

### Running the Application

1. **Regular Scraping**:
   ```bash
   go run main.go
   ```

2. **Generate Markdown Report**:
   ```bash
   go run main.go -markdown
   ```

3. **Build and Run**:
   ```bash
   go build -o scraper
   ./scraper
   ```

## Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/usecases/
```

## Configuration Options

The application supports various configuration options through environment variables:

### Database Configuration
- `DATABASE_URL`: Full PostgreSQL connection string
- `DB_MAX_CONNS`: Maximum database connections (default: 25)
- `DB_MIN_CONNS`: Minimum database connections (default: 5)

### Application Configuration  
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `ENVIRONMENT`: Environment (development, staging, production)
- `GRACEFUL_TIMEOUT`: Shutdown timeout (default: 30s)
- `HACKERNEWS_CURRENT_ID`: Current HackerNews post ID
- `MAX_JOBS_TO_GET`: Maximum jobs per source

### Scraper Configuration
- RSS feeds are configured in the container setup
- Reddit subreddits: `forhire`, `remotejs`
- HackerNews: "Who is hiring?" posts

## Legacy Code

The original implementation has been preserved in:
- `main_legacy.go`: Original main function
- `pkg/`: Original package structure

This allows for easy comparison and gradual migration if needed.

## Database Management

Using Drizzle ORM for schema management:

- **Generate migrations**: `pnpm db:generate`
- **Apply migrations**: `pnpm db:migrate`  
- **Push schema changes**: `pnpm db:push` (development only)
- **Open database studio**: `pnpm db:studio`

## Architecture Benefits

### Maintainability
- Clear separation of concerns
- Single responsibility principle
- Easy to locate and modify functionality

### Testability
- Interface-driven design enables mocking
- Isolated business logic testing
- Comprehensive test coverage

### Scalability  
- Pluggable scraper architecture
- Concurrent processing
- Connection pooling and resource management

### Reliability
- Proper error handling and recovery
- Graceful shutdown and cleanup
- Transaction support for data consistency

### Observability
- Structured logging throughout
- Component-based logging
- Error context and tracing

## Contributing

1. Follow Go best practices and conventions
2. Add tests for new functionality
3. Update documentation as needed
4. Use structured logging for observability
5. Follow the established architecture patterns

## Migration from Legacy

The new architecture maintains backward compatibility while providing:
- Better error handling and logging
- Improved configuration management
- Enhanced testability
- Proper resource management
- Clean separation of concerns

The legacy code remains available for reference and gradual migration.

## License
Distributed under the MIT License. See `LICENSE` for more information.
package repositories

import (
	"context"
	"time"

	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
)

// NoticeRepository defines the interface for notice data operations
type NoticeRepository interface {
	// Create inserts new notices, handling duplicates gracefully
	Create(ctx context.Context, notices []*entities.Notice) error
	
	// GetLatest retrieves the most recent notices up to the specified limit
	GetLatest(ctx context.Context, limit int) ([]*entities.Notice, error)
	
	// GetCount returns the total number of notices in the repository
	GetCount(ctx context.Context) (int, error)
	
	// GetLatestByDate retrieves notices created within the specified timeframe
	GetLatestByDate(ctx context.Context, since time.Time) ([]*entities.Notice, error)
	
	// GetBySource retrieves notices from a specific source
	GetBySource(ctx context.Context, sourceID string, limit int) ([]*entities.Notice, error)
}
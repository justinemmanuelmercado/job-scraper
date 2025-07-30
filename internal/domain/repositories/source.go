package repositories

import (
	"context"

	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
)

// SourceRepository defines the interface for source data operations
type SourceRepository interface {
	// GetByName retrieves a source by its name
	GetByName(ctx context.Context, name string) (*entities.Source, error)
	
	// Create creates a new source
	Create(ctx context.Context, source *entities.Source) error
	
	// List retrieves all sources
	List(ctx context.Context) ([]*entities.Source, error)
}
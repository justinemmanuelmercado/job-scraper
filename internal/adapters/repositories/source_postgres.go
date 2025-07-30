package repositories

import (
	"context"
	"database/sql"

	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/repositories"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/errors"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/logging"
)

// PostgresSourceRepository implements SourceRepository using PostgreSQL
type PostgresSourceRepository struct {
	db     *sql.DB
	logger *logging.Logger
}

// NewPostgresSourceRepository creates a new PostgreSQL source repository
func NewPostgresSourceRepository(db *sql.DB, logger *logging.Logger) repositories.SourceRepository {
	return &PostgresSourceRepository{
		db:     db,
		logger: logger,
	}
}

// GetByName retrieves a source by its name
func (r *PostgresSourceRepository) GetByName(ctx context.Context, name string) (*entities.Source, error) {
	query := `
		SELECT name, description, "createdAt", "updatedAt", homepage, "iconUrl"
		FROM "Source"
		WHERE name = $1
	`

	source := &entities.Source{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&source.Name,
		&source.Description,
		&source.CreatedAt,
		&source.UpdatedAt,
		&source.Homepage,
		&source.IconURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDatabaseError("source not found").WithContext("name", name)
		}
		return nil, errors.WrapDatabaseError(err, "failed to get source by name")
	}

	return source, nil
}

// Create creates a new source
func (r *PostgresSourceRepository) Create(ctx context.Context, source *entities.Source) error {
	query := `
		INSERT INTO "Source" (name, description, homepage, "iconUrl")
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query,
		source.Name,
		source.Description,
		source.Homepage,
		source.IconURL,
	)

	if err != nil {
		return errors.WrapDatabaseError(err, "failed to create source")
	}

	r.logger.WithFields(map[string]interface{}{
		"source_name": source.Name,
	}).Info("Successfully created source")

	return nil
}

// List retrieves all sources
func (r *PostgresSourceRepository) List(ctx context.Context) ([]*entities.Source, error) {
	query := `
		SELECT name, description, "createdAt", "updatedAt", homepage, "iconUrl"
		FROM "Source"
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.WrapDatabaseError(err, "failed to list sources")
	}
	defer rows.Close()

	var sources []*entities.Source
	for rows.Next() {
		source := &entities.Source{}
		err := rows.Scan(
			&source.Name,
			&source.Description,
			&source.CreatedAt,
			&source.UpdatedAt,
			&source.Homepage,
			&source.IconURL,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan source row")
			continue
		}
		sources = append(sources, source)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapDatabaseError(err, "error occurred during row iteration")
	}

	return sources, nil
}
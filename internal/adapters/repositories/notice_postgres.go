package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/repositories"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/errors"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/logging"
)

// PostgresNoticeRepository implements NoticeRepository using PostgreSQL
type PostgresNoticeRepository struct {
	db     *sql.DB
	logger *logging.Logger
}

// NewPostgresNoticeRepository creates a new PostgreSQL notice repository
func NewPostgresNoticeRepository(db *sql.DB, logger *logging.Logger) repositories.NoticeRepository {
	return &PostgresNoticeRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts new notices, handling duplicates gracefully
func (r *PostgresNoticeRepository) Create(ctx context.Context, notices []*entities.Notice) error {
	if len(notices) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WrapDatabaseError(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	query := `
		INSERT INTO "Notice" (
			id, title, body, url, "authorName", "authorUrl", "imageUrl",
			"sourceId", raw, guid, "publishedDate"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (guid, "sourceId") DO NOTHING
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return errors.WrapDatabaseError(err, "failed to prepare statement")
	}
	defer stmt.Close()

	insertedCount := 0
	for _, notice := range notices {
		result, err := stmt.ExecContext(ctx,
			notice.ID,
			notice.Title,
			notice.Body,
			notice.URL,
			notice.AuthorName,
			notice.AuthorURL,
			notice.ImageURL,
			notice.SourceID,
			notice.Raw,
			notice.Guid,
			notice.PublishedDate,
		)
		if err != nil {
			r.logger.WithError(err).WithFields(map[string]interface{}{
				"notice_id": notice.ID,
				"source":    notice.SourceID,
			}).Error("Failed to insert notice")
			continue
		}

		if affected, _ := result.RowsAffected(); affected > 0 {
			insertedCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.WrapDatabaseError(err, "failed to commit transaction")
	}

	r.logger.WithFields(map[string]interface{}{
		"total_notices":    len(notices),
		"inserted_notices": insertedCount,
	}).Info("Successfully created notices")

	return nil
}

// GetLatest retrieves the most recent notices up to the specified limit
func (r *PostgresNoticeRepository) GetLatest(ctx context.Context, limit int) ([]*entities.Notice, error) {
	query := `
		SELECT id, title, body, url, "authorName", "authorUrl", "imageUrl",
			   "createdAt", "updatedAt", "sourceId", raw, guid, "publishedDate"
		FROM "Notice"
		ORDER BY "createdAt" DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, errors.WrapDatabaseError(err, "failed to query latest notices")
	}
	defer rows.Close()

	var notices []*entities.Notice
	for rows.Next() {
		notice := &entities.Notice{}
		err := rows.Scan(
			&notice.ID,
			&notice.Title,
			&notice.Body,
			&notice.URL,
			&notice.AuthorName,
			&notice.AuthorURL,
			&notice.ImageURL,
			&notice.CreatedAt,
			&notice.UpdatedAt,
			&notice.SourceID,
			&notice.Raw,
			&notice.Guid,
			&notice.PublishedDate,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan notice row")
			continue
		}
		notices = append(notices, notice)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapDatabaseError(err, "error occurred during row iteration")
	}

	return notices, nil
}

// GetCount returns the total number of notices in the repository
func (r *PostgresNoticeRepository) GetCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM "Notice"`
	
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, errors.WrapDatabaseError(err, "failed to get notice count")
	}

	return count, nil
}

// GetLatestByDate retrieves notices created within the specified timeframe
func (r *PostgresNoticeRepository) GetLatestByDate(ctx context.Context, since time.Time) ([]*entities.Notice, error) {
	query := `
		SELECT id, title, body, url, "authorName", "authorUrl", "imageUrl",
			   "createdAt", "updatedAt", "sourceId", raw, guid, "publishedDate"
		FROM "Notice"
		WHERE "createdAt" >= $1
		ORDER BY "publishedDate" DESC
	`

	rows, err := r.db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, errors.WrapDatabaseError(err, "failed to query notices by date")
	}
	defer rows.Close()

	var notices []*entities.Notice
	for rows.Next() {
		notice := &entities.Notice{}
		err := rows.Scan(
			&notice.ID,
			&notice.Title,
			&notice.Body,
			&notice.URL,
			&notice.AuthorName,
			&notice.AuthorURL,
			&notice.ImageURL,
			&notice.CreatedAt,
			&notice.UpdatedAt,
			&notice.SourceID,
			&notice.Raw,
			&notice.Guid,
			&notice.PublishedDate,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan notice row")
			continue
		}
		notices = append(notices, notice)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapDatabaseError(err, "error occurred during row iteration")
	}

	return notices, nil
}

// GetBySource retrieves notices from a specific source
func (r *PostgresNoticeRepository) GetBySource(ctx context.Context, sourceID string, limit int) ([]*entities.Notice, error) {
	query := `
		SELECT id, title, body, url, "authorName", "authorUrl", "imageUrl",
			   "createdAt", "updatedAt", "sourceId", raw, guid, "publishedDate"
		FROM "Notice"
		WHERE "sourceId" = $1
		ORDER BY "createdAt" DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, sourceID, limit)
	if err != nil {
		return nil, errors.WrapDatabaseError(err, fmt.Sprintf("failed to query notices for source %s", sourceID))
	}
	defer rows.Close()

	var notices []*entities.Notice
	for rows.Next() {
		notice := &entities.Notice{}
		err := rows.Scan(
			&notice.ID,
			&notice.Title,
			&notice.Body,
			&notice.URL,
			&notice.AuthorName,
			&notice.AuthorURL,
			&notice.ImageURL,
			&notice.CreatedAt,
			&notice.UpdatedAt,
			&notice.SourceID,
			&notice.Raw,
			&notice.Guid,
			&notice.PublishedDate,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan notice row")
			continue
		}
		notices = append(notices, notice)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapDatabaseError(err, "error occurred during row iteration")
	}

	return notices, nil
}
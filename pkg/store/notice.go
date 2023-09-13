package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	errorHandler "github.com/justinemmanuelmercado/go-scraper/pkg"
	"github.com/justinemmanuelmercado/go-scraper/pkg/models"
)

const tableName = "Notice"

type NoticeStore struct {
	conn *pgx.Conn
}

func InitNotice(conn *pgx.Conn) *NoticeStore {
	return &NoticeStore{conn: conn}
}

func (n *NoticeStore) CreateNotices(notices []*models.Notice) error {
	query := fmt.Sprintf(`
	INSERT INTO "%s" (
		id,
		title,
		body,
		url,
		"authorName",
		"authorUrl",
		"imageUrl",
		"sourceId",
		raw,
		guid,
		"publishedDate"
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11
	) ON CONFLICT (guid, "sourceId") DO NOTHING`, tableName)

	batch := &pgx.Batch{}

	for _, notice := range notices {
		batch.Queue(
			query,
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
	}

	br := n.conn.SendBatch(context.Background(), batch)
	_, err := br.Exec()
	if err != nil {
		return err
	}
	return br.Close()
}

func (n *NoticeStore) GetCount() int {
	var count int
	err := n.conn.QueryRow(context.Background(), fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tableName)).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (n *NoticeStore) GetLatest(count int) []models.Notice {
	var notices []models.Notice
	rows, err := n.conn.Query(context.Background(), fmt.Sprintf(`SELECT * FROM "%s" ORDER BY "createdAt" DESC LIMIT $1`, tableName), count)
	if err != nil {
		errorHandler.HandleErrorWithSection(err, "Error querying latest notices", "Database")
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var notice models.Notice
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
			errorHandler.HandleErrorWithSection(err, "Error scanning row", "Database")
			return nil
		}
		notices = append(notices, notice)
	}

	err = rows.Err()
	if err != nil {
		errorHandler.HandleErrorWithSection(err, "Error iterating rows", "Database")
		return nil
	}

	return notices
}

func (n *NoticeStore) GetLatestNotices() ([]*models.Notice, error) {
	rows, err := n.conn.Query(context.Background(), `
	SELECT * FROM "Notice"
	WHERE "createdAt" >= (now() - interval '1 day')
	AND "sourceId" != 'Reddit'
	ORDER BY "publishedDate" DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notices := []*models.Notice{}
	for rows.Next() {
		var notice models.Notice
		// Assume models.Notice has the same fields as your Prisma schema
		if err := rows.Scan(&notice.ID,
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
			&notice.PublishedDate); err != nil {
			return nil, err
		}
		notices = append(notices, &notice)
	}

	return notices, nil
}

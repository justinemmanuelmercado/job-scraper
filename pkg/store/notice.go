package store

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/justinemmanuelmercado/go-scraper/pkg/models"
)

type Notice struct {
	conn *pgx.Conn
}

func InitNotice(conn *pgx.Conn) *Notice {
	return &Notice{conn: conn}
}

func (n *Notice) CreateNotices(notices []*models.Notice) error {
	query := `
	INSERT INTO "Notice" (id, title, body, url, "authorName", "authorUrl", "imageUrl", "sourceId", raw, guid, "publishedDate")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	ON CONFLICT (guid, "sourceId") DO NOTHING`

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

func (n *Notice) GetNotice(id string) (*models.Notice, error) {
	var notice models.Notice
	err := n.conn.QueryRow(context.Background(), `SELECT * FROM "Notice" WHERE id=$1`, id).Scan(
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
	)

	if err != nil {
		return nil, err
	}

	return &notice, nil
}

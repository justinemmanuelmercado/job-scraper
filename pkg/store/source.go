package store

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/justinemmanuelmercado/go-scraper/pkg/models"
)

type Source struct {
	conn *pgx.Conn
}

func InitSource(conn *pgx.Conn) *Source {
	return &Source{conn}
}

func (s *Source) GetSourceByName(name string) (*models.Source, error) {
	var source *models.Source = &models.Source{}
	err := s.conn.QueryRow(context.Background(), `SELECT id, name, description FROM "Source" WHERE "name" = $1`, name).Scan(
		&source.ID,
		&source.Name,
		&source.Description,
	)

	if err != nil {
		return nil, err
	}

	return source, nil
}

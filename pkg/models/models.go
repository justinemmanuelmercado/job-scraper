package models

import (
	"time"
)

type Notice struct {
	ID            string
	Title         string
	Body          string
	URL           string
	AuthorName    string
	AuthorURL     string
	ImageURL      *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	SourceID      string
	Raw           string
	Guid          string
	PublishedDate *time.Time
}

type Keyword struct {
	ID        string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Source struct {
	ID          string
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

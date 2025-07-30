package entities

import (
	"time"
)

// Notice represents a job posting from any source
type Notice struct {
	ID            string
	Title         string
	Body          string
	URL           string
	AuthorName    string
	AuthorURL     string
	ImageURL      *string
	CreatedAt     time.Time
	UpdatedAt     *time.Time
	SourceID      string
	Raw           string
	Guid          string
	PublishedDate *time.Time
}

// Keyword represents a searchable keyword
type Keyword struct {
	ID        string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Source represents a job source (RSS feed, Reddit, etc.)
type Source struct {
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Homepage    *string
	IconURL     *string
}

// ScrapeResult contains the results of a scraping operation
type ScrapeResult struct {
	SourceName     string
	Notices        []*Notice
	ScrapedCount   int
	NewCount       int
	ErrorMessages  []string
}

// Summary contains overall scraping statistics
type Summary struct {
	TotalScraped   int
	TotalNew       int
	Duration       time.Duration
	Sources        []ScrapeResult
	Timestamp      time.Time
}
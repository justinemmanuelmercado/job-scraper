package services

import (
	"context"

	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
)

// Scraper defines the interface for job scrapers
type Scraper interface {
	// Name returns the human-readable name of the scraper
	Name() string
	
	// Scrape performs the scraping operation and returns notices
	Scrape(ctx context.Context) ([]*entities.Notice, error)
	
	// IsEnabled returns whether this scraper is currently enabled
	IsEnabled() bool
}

// ScrapingService orchestrates multiple scrapers
type ScrapingService interface {
	// ScrapeAll runs all enabled scrapers concurrently
	ScrapeAll(ctx context.Context) ([]*entities.ScrapeResult, error)
	
	// RegisterScraper adds a new scraper to the service
	RegisterScraper(scraper Scraper)
	
	// GetScrapers returns all registered scrapers
	GetScrapers() []Scraper
}
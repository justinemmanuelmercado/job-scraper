package scrapers

import (
	"context"
	"sync"
	"time"

	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/services"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/logging"
)

// ScrapingServiceImpl implements the ScrapingService interface
type ScrapingServiceImpl struct {
	scrapers []services.Scraper
	logger   *logging.Logger
}

// NewScrapingService creates a new scraping service
func NewScrapingService(logger *logging.Logger) services.ScrapingService {
	return &ScrapingServiceImpl{
		scrapers: make([]services.Scraper, 0),
		logger:   logger,
	}
}

// RegisterScraper adds a new scraper to the service
func (s *ScrapingServiceImpl) RegisterScraper(scraper services.Scraper) {
	s.scrapers = append(s.scrapers, scraper)
	s.logger.WithFields(map[string]interface{}{
		"scraper_name": scraper.Name(),
		"enabled":      scraper.IsEnabled(),
	}).Info("Registered scraper")
}

// GetScrapers returns all registered scrapers
func (s *ScrapingServiceImpl) GetScrapers() []services.Scraper {
	return s.scrapers
}

// ScrapeAll runs all enabled scrapers concurrently
func (s *ScrapingServiceImpl) ScrapeAll(ctx context.Context) ([]*entities.ScrapeResult, error) {
	enabledScrapers := s.getEnabledScrapers()
	
	if len(enabledScrapers) == 0 {
		s.logger.Warn("No enabled scrapers found")
		return []*entities.ScrapeResult{}, nil
	}

	s.logger.WithFields(map[string]interface{}{
		"scraper_count": len(enabledScrapers),
	}).Info("Starting concurrent scraping")

	var wg sync.WaitGroup
	resultsCh := make(chan *entities.ScrapeResult, len(enabledScrapers))

	// Run each scraper concurrently
	for _, scraper := range enabledScrapers {
		wg.Add(1)
		go s.runScraper(ctx, scraper, &wg, resultsCh)
	}

	wg.Wait()
	close(resultsCh)

	// Collect results
	var allResults []*entities.ScrapeResult
	for result := range resultsCh {
		allResults = append(allResults, result)
	}

	s.logger.WithFields(map[string]interface{}{
		"results_count": len(allResults),
	}).Info("Completed concurrent scraping")

	return allResults, nil
}

// getEnabledScrapers returns only the enabled scrapers
func (s *ScrapingServiceImpl) getEnabledScrapers() []services.Scraper {
	var enabled []services.Scraper
	for _, scraper := range s.scrapers {
		if scraper.IsEnabled() {
			enabled = append(enabled, scraper)
		}
	}
	return enabled
}

// runScraper runs a single scraper and handles its result
func (s *ScrapingServiceImpl) runScraper(ctx context.Context, scraper services.Scraper, wg *sync.WaitGroup, resultsCh chan<- *entities.ScrapeResult) {
	defer wg.Done()

	scraperLogger := s.logger.WithFields(map[string]interface{}{
		"scraper_name": scraper.Name(),
	})

	startTime := time.Now()
	scraperLogger.Info("Starting scraper")

	// Check context cancellation
	select {
	case <-ctx.Done():
		scraperLogger.Warn("Context cancelled, skipping scraper")
		resultsCh <- &entities.ScrapeResult{
			SourceName:    scraper.Name(),
			Notices:       []*entities.Notice{},
			ScrapedCount:  0,
			NewCount:      0,
			ErrorMessages: []string{"context cancelled"},
		}
		return
	default:
	}

	// Run the scraper
	notices, err := scraper.Scrape(ctx)
	duration := time.Since(startTime)

	result := &entities.ScrapeResult{
		SourceName:   scraper.Name(),
		Notices:      notices,
		ScrapedCount: len(notices),
		NewCount:     0, // This will be updated after database insertion
	}

	if err != nil {
		result.ErrorMessages = []string{err.Error()}
		scraperLogger.WithError(err).WithFields(map[string]interface{}{
			"duration": duration,
		}).Error("Scraper failed")
	} else {
		scraperLogger.WithFields(map[string]interface{}{
			"notices_count": len(notices),
			"duration":      duration,
		}).Info("Scraper completed successfully")
	}

	resultsCh <- result
}
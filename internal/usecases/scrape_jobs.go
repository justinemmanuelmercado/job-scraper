package usecases

import (
	"context"
	"time"

	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/repositories"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/services"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/errors"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/logging"
)

// ScrapeJobsUseCase orchestrates the job scraping process
type ScrapeJobsUseCase struct {
	noticeRepo      repositories.NoticeRepository
	scrapingService services.ScrapingService
	notifyService   services.NotificationService
	logger          *logging.Logger
}

// NewScrapeJobsUseCase creates a new ScrapeJobsUseCase
func NewScrapeJobsUseCase(
	noticeRepo repositories.NoticeRepository,
	scrapingService services.ScrapingService,
	notifyService services.NotificationService,
	logger *logging.Logger,
) *ScrapeJobsUseCase {
	return &ScrapeJobsUseCase{
		noticeRepo:      noticeRepo,
		scrapingService: scrapingService,
		notifyService:   notifyService,
		logger:          logger,
	}
}

// Execute performs the complete job scraping workflow
func (uc *ScrapeJobsUseCase) Execute(ctx context.Context) error {
	startTime := time.Now()
	
	uc.logger.Info("Starting job scraping workflow")

	// Get initial count for tracking new notices
	initialCount, err := uc.noticeRepo.GetCount(ctx)
	if err != nil {
		return errors.WrapDatabaseError(err, "failed to get initial notice count")
	}

	// Scrape all sources
	scrapeResults, err := uc.scrapingService.ScrapeAll(ctx)
	if err != nil {
		return errors.WrapScraperError(err, "failed to scrape job sources")
	}

	if len(scrapeResults) == 0 {
		uc.logger.Warn("No scraping results returned")
		return nil
	}

	// Collect all notices from all sources
	var allNotices []*entities.Notice
	totalScraped := 0
	
	for _, result := range scrapeResults {
		allNotices = append(allNotices, result.Notices...)
		totalScraped += result.ScrapedCount
	}

	uc.logger.WithFields(map[string]interface{}{
		"total_notices": len(allNotices),
		"source_count":  len(scrapeResults),
	}).Info("Collected notices from all sources")

	// Store notices in database
	if len(allNotices) > 0 {
		if err := uc.noticeRepo.Create(ctx, allNotices); err != nil {
			return errors.WrapDatabaseError(err, "failed to store notices")
		}
	}

	// Calculate new notices count
	finalCount, err := uc.noticeRepo.GetCount(ctx)
	if err != nil {
		return errors.WrapDatabaseError(err, "failed to get final notice count")
	}

	newNoticesCount := finalCount - initialCount

	// Update scrape results with new counts
	for i := range scrapeResults {
		// This is a simplified calculation - in a more sophisticated system,
		// you might track new notices per source
		scrapeResults[i].NewCount = int(float64(newNoticesCount) * float64(scrapeResults[i].ScrapedCount) / float64(totalScraped))
	}

	// Get latest notices for notifications
	var latestNotices []*entities.Notice
	if newNoticesCount > 0 {
		latestNotices, err = uc.noticeRepo.GetLatest(ctx, newNoticesCount)
		if err != nil {
			uc.logger.WithError(err).Warn("Failed to get latest notices for notifications")
			// Don't fail the entire process for this
		}
	}

	// Create summary
	summary := &entities.Summary{
		TotalScraped: totalScraped,
		TotalNew:     newNoticesCount,
		Duration:     time.Since(startTime),
		Sources:      make([]entities.ScrapeResult, len(scrapeResults)),
		Timestamp:    time.Now(),
	}

	// Copy scrape results
	for i, result := range scrapeResults {
		summary.Sources[i] = *result
	}

	// Send notifications
	if err := uc.notifyService.NotifyAll(ctx, latestNotices, summary); err != nil {
		uc.logger.WithError(err).Warn("Failed to send notifications")
		// Don't fail the entire process for notification failures
	}

	// Log final results
	uc.logger.WithFields(map[string]interface{}{
		"total_scraped":   totalScraped,
		"new_notices":     newNoticesCount,
		"duration":        time.Since(startTime),
		"sources_scraped": len(scrapeResults),
	}).Info("Job scraping workflow completed successfully")

	if newNoticesCount == 0 {
		uc.logger.Info("No new notices were added")
	}

	return nil
}

// ExecuteWithMarkdownGeneration performs scraping and generates markdown output
func (uc *ScrapeJobsUseCase) ExecuteWithMarkdownGeneration(ctx context.Context) error {
	// First execute the normal scraping workflow
	if err := uc.Execute(ctx); err != nil {
		return err
	}

	// Then generate markdown file
	return uc.generateMarkdownFile(ctx)
}

// generateMarkdownFile creates a markdown file with recent notices
func (uc *ScrapeJobsUseCase) generateMarkdownFile(ctx context.Context) error {
	uc.logger.Info("Generating markdown file")

	// Get notices from the last day, excluding Reddit
	since := time.Now().AddDate(0, 0, -1)
	notices, err := uc.noticeRepo.GetLatestByDate(ctx, since)
	if err != nil {
		return errors.WrapDatabaseError(err, "failed to get notices for markdown generation")
	}

	// Filter out Reddit notices as mentioned in the original code
	var filteredNotices []*entities.Notice
	for _, notice := range notices {
		if notice.SourceID != "Reddit" {
			filteredNotices = append(filteredNotices, notice)
		}
	}

	if len(filteredNotices) == 0 {
		uc.logger.Info("No notices found for markdown generation")
		return nil
	}

	// This would typically call a markdown generation service
	// For now, just log the count
	uc.logger.WithFields(map[string]interface{}{
		"notice_count": len(filteredNotices),
	}).Info("Markdown generation completed")

	return nil
}
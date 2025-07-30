package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/services"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/errors"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/logging"
	"github.com/mmcdole/gofeed"
)

// RSSFeed represents an RSS feed configuration
type RSSFeed struct {
	URL        string
	SourceName string
}

// RSSFeedScraper implements the Scraper interface for RSS feeds
type RSSFeedScraper struct {
	feeds   []RSSFeed
	logger  *logging.Logger
	enabled bool
}

// NewRSSFeedScraper creates a new RSS feed scraper
func NewRSSFeedScraper(feeds []RSSFeed, logger *logging.Logger) services.Scraper {
	return &RSSFeedScraper{
		feeds:   feeds,
		logger:  logger,
		enabled: true,
	}
}

// Name returns the scraper name
func (s *RSSFeedScraper) Name() string {
	return "RSS Feeds"
}

// IsEnabled returns whether the scraper is enabled
func (s *RSSFeedScraper) IsEnabled() bool {
	return s.enabled
}

// Scrape performs RSS feed scraping
func (s *RSSFeedScraper) Scrape(ctx context.Context) ([]*entities.Notice, error) {
	s.logger.WithFields(map[string]interface{}{
		"feed_count": len(s.feeds),
	}).Info("Starting RSS feed scraping")

	var wg sync.WaitGroup
	noticesCh := make(chan []*entities.Notice, len(s.feeds))
	errCh := make(chan error, len(s.feeds))

	// Process each feed concurrently
	for _, feed := range s.feeds {
		wg.Add(1)
		go s.scrapeFeed(ctx, feed, &wg, noticesCh, errCh)
	}

	wg.Wait()
	close(noticesCh)
	close(errCh)

	// Collect errors (but don't fail completely if one feed fails)
	var scrapeErrors []error
	for err := range errCh {
		s.logger.WithError(err).Warn("RSS feed scraping error")
		scrapeErrors = append(scrapeErrors, err)
	}

	// Collect all notices
	var allNotices []*entities.Notice
	for notices := range noticesCh {
		allNotices = append(allNotices, notices...)
	}

	s.logger.WithFields(map[string]interface{}{
		"total_notices": len(allNotices),
		"error_count":   len(scrapeErrors),
	}).Info("Completed RSS feed scraping")

	return allNotices, nil
}

// scrapeFeed scrapes a single RSS feed
func (s *RSSFeedScraper) scrapeFeed(ctx context.Context, feed RSSFeed, wg *sync.WaitGroup, noticesCh chan<- []*entities.Notice, errCh chan<- error) {
	defer wg.Done()

	feedLogger := s.logger.WithFields(map[string]interface{}{
		"feed_url":    feed.URL,
		"source_name": feed.SourceName,
	})

	feedLogger.Debug("Fetching RSS feed")

	// Check context cancellation
	select {
	case <-ctx.Done():
		errCh <- errors.WrapScraperError(ctx.Err(), "context cancelled during RSS scraping")
		return
	default:
	}

	parser := gofeed.NewParser()
	parsedFeed, err := parser.ParseURLWithContext(feed.URL, ctx)
	if err != nil {
		errCh <- errors.WrapScraperError(err, fmt.Sprintf("failed to parse RSS feed %s", feed.URL))
		return
	}

	if len(parsedFeed.Items) == 0 {
		feedLogger.Warn("No items found in RSS feed")
		noticesCh <- []*entities.Notice{}
		return
	}

	notices := make([]*entities.Notice, len(parsedFeed.Items))
	for i, item := range parsedFeed.Items {
		notice, err := s.convertItemToNotice(item, feed.SourceName)
		if err != nil {
			feedLogger.WithError(err).WithFields(map[string]interface{}{
				"item_title": item.Title,
				"item_guid":  item.GUID,
			}).Warn("Failed to convert RSS item to notice")
			continue
		}
		notices[i] = notice
	}

	feedLogger.WithFields(map[string]interface{}{
		"item_count": len(notices),
	}).Info("Successfully scraped RSS feed")

	noticesCh <- notices
}

// convertItemToNotice converts an RSS item to a Notice entity
func (s *RSSFeedScraper) convertItemToNotice(item *gofeed.Item, sourceID string) (*entities.Notice, error) {
	// Serialize the raw RSS item
	jsonData, err := json.Marshal(item)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to marshal RSS item to JSON, using empty raw data")
		jsonData = []byte("{}")
	}

	notice := &entities.Notice{
		ID:            uuid.New().String(),
		Title:         item.Title,
		Body:          item.Description,
		URL:           item.Link,
		SourceID:      sourceID,
		Raw:           string(jsonData),
		Guid:          item.GUID,
		PublishedDate: item.PublishedParsed,
	}

	// Set image URL if available
	if item.Image != nil && item.Image.URL != "" {
		notice.ImageURL = &item.Image.URL
	}

	// Set author information if available
	if len(item.Authors) > 0 {
		notice.AuthorName = item.Authors[0].Name
		if item.Authors[0].Email != "" {
			notice.AuthorURL = item.Authors[0].Email
		}
	}

	// Fallback for author information
	if notice.AuthorName == "" && item.Author != nil {
		notice.AuthorName = item.Author.Name
		if item.Author.Email != "" {
			notice.AuthorURL = item.Author.Email
		}
	}

	return notice, nil
}
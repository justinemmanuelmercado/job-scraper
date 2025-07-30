package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/services"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/config"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/errors"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/logging"
	"github.com/thecsw/mira"
)

// RedditClient interface for easier testing
type RedditClient interface {
	GetSubredditPosts(sr string, sort string, duration string, limit int) ([]mira.PostListingChild, error)
}

// RedditScraper implements the Scraper interface for Reddit
type RedditScraper struct {
	client     RedditClient
	subreddits []string
	logger     *logging.Logger
	enabled    bool
}

// NewRedditScraper creates a new Reddit scraper
func NewRedditScraper(cfg config.RedditConfig, logger *logging.Logger) services.Scraper {
	client, err := mira.Init(mira.Credentials{
		ClientId:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Username:     cfg.Username,
		Password:     cfg.Password,
		UserAgent:   cfg.UserAgent,
	})

	if err != nil {
		logger.WithError(err).Error("Failed to initialize Reddit client")
		return &RedditScraper{enabled: false, logger: logger}
	}

	return &RedditScraper{
		client:     client,
		subreddits: cfg.Subreddits,
		logger:     logger,
		enabled:    true,
	}
}

// Name returns the scraper name
func (s *RedditScraper) Name() string {
	return "Reddit"
}

// IsEnabled returns whether the scraper is enabled
func (s *RedditScraper) IsEnabled() bool {
	return s.enabled
}

// Scrape performs Reddit scraping
func (s *RedditScraper) Scrape(ctx context.Context) ([]*entities.Notice, error) {
	if !s.enabled {
		return nil, errors.NewScraperError("Reddit scraper is disabled")
	}

	s.logger.WithFields(map[string]interface{}{
		"subreddit_count": len(s.subreddits),
	}).Info("Starting Reddit scraping")

	var wg sync.WaitGroup
	postsCh := make(chan []mira.PostListingChild, len(s.subreddits))
	errCh := make(chan error, len(s.subreddits))

	// Scrape each subreddit concurrently
	for _, subreddit := range s.subreddits {
		wg.Add(1)
		go s.scrapeSubreddit(ctx, subreddit, &wg, postsCh, errCh)
	}

	wg.Wait()
	close(postsCh)
	close(errCh)

	// Check for errors
	if len(errCh) > 0 {
		err := <-errCh
		s.logger.WithError(err).Error("Reddit scraping failed")
		return nil, errors.WrapScraperError(err, "failed to scrape Reddit")
	}

	// Collect all posts
	var allPosts []mira.PostListingChild
	for posts := range postsCh {
		allPosts = append(allPosts, posts...)
	}

	// Convert posts to notices
	notices := make([]*entities.Notice, len(allPosts))
	for i, post := range allPosts {
		notice, err := s.convertPostToNotice(post)
		if err != nil {
			s.logger.WithError(err).WithFields(map[string]interface{}{
				"post_id":    post.Data.Id,
				"post_title": post.Data.Title,
			}).Warn("Failed to convert Reddit post to notice")
			continue
		}
		notices[i] = notice
	}

	s.logger.WithFields(map[string]interface{}{
		"total_notices": len(notices),
	}).Info("Completed Reddit scraping")

	return notices, nil
}

// scrapeSubreddit scrapes a single subreddit
func (s *RedditScraper) scrapeSubreddit(ctx context.Context, subreddit string, wg *sync.WaitGroup, postsCh chan<- []mira.PostListingChild, errCh chan<- error) {
	defer wg.Done()

	subredditLogger := s.logger.WithFields(map[string]interface{}{
		"subreddit": subreddit,
	})

	// Check context cancellation
	select {
	case <-ctx.Done():
		errCh <- errors.WrapScraperError(ctx.Err(), "context cancelled during Reddit scraping")
		return
	default:
	}

	subredditLogger.Debug("Fetching posts from subreddit")

	posts, err := s.client.GetSubredditPosts(subreddit, "new", "week", 20)
	if err != nil {
		errCh <- errors.WrapScraperError(err, fmt.Sprintf("failed to fetch posts from subreddit %s", subreddit))
		return
	}

	// Filter for hiring posts
	filteredPosts := s.filterHiringPosts(posts)

	subredditLogger.WithFields(map[string]interface{}{
		"total_posts":    len(posts),
		"filtered_posts": len(filteredPosts),
	}).Info("Successfully scraped subreddit")

	postsCh <- filteredPosts
}

// filterHiringPosts filters posts that are likely hiring posts
func (s *RedditScraper) filterHiringPosts(posts []mira.PostListingChild) []mira.PostListingChild {
	var hiringPosts []mira.PostListingChild
	
	for _, post := range posts {
		title := post.Data.Title
		if s.isHiringPost(title) {
			hiringPosts = append(hiringPosts, post)
		}
	}
	
	return hiringPosts
}

// isHiringPost determines if a post is a hiring post based on title
func (s *RedditScraper) isHiringPost(title string) bool {
	// This is a simplified version - you might want to use the original filter logic
	hiringKeywords := []string{"hiring", "job", "remote", "developer", "engineer", "programmer"}
	
	titleLower := strings.ToLower(title)
	for _, keyword := range hiringKeywords {
		if strings.Contains(titleLower, keyword) {
			return true
		}
	}
	
	return false
}

// convertPostToNotice converts a Reddit post to a Notice entity
func (s *RedditScraper) convertPostToNotice(post mira.PostListingChild) (*entities.Notice, error) {
	// Serialize the raw post data
	jsonData, err := json.Marshal(post)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to marshal Reddit post to JSON, using empty raw data")
		jsonData = []byte("{}")
	}

	publishedTime := time.Unix(int64(post.GetTimeCreated()), 0).UTC()

	notice := &entities.Notice{
		ID:            uuid.New().String(),
		Title:         post.Data.Title,
		Body:          html.UnescapeString(post.Data.SelftextHtml),
		URL:           fmt.Sprintf("https://www.reddit.com/%s", post.GetPermalink()),
		AuthorName:    post.GetAuthor(),
		AuthorURL:     fmt.Sprintf("https://www.reddit.com/user/%s", post.GetAuthor()),
		SourceID:      "Reddit",
		Raw:           string(jsonData),
		Guid:          post.Data.Id,
		PublishedDate: &publishedTime,
	}

	// Set image URL if available
	if len(post.Data.Preview.Images) > 0 {
		notice.ImageURL = &post.Data.Preview.Images[0].Source.Url
	}

	return notice, nil
}
package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/services"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/errors"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/logging"
)

// HackerNewsStory represents a HackerNews story
type HackerNewsStory struct {
	Time int64  `json:"time"`
	Text string `json:"text"`
	ID   int    `json:"id"`
	By   string `json:"by"`
	Raw  string `json:"raw"`
}

// HackerNewsParent represents a HackerNews parent post with kids
type HackerNewsParent struct {
	Kids []int `json:"kids"`
	Time int64  `json:"time"`
}

// HackerNewsScraper implements the Scraper interface for HackerNews
type HackerNewsScraper struct {
	currentID    int
	maxJobsToGet int
	logger       *logging.Logger
	enabled      bool
	httpClient   *http.Client
}

// NewHackerNewsScraper creates a new HackerNews scraper
func NewHackerNewsScraper(currentID, maxJobsToGet int, logger *logging.Logger) services.Scraper {
	return &HackerNewsScraper{
		currentID:    currentID,
		maxJobsToGet: maxJobsToGet,
		logger:       logger,
		enabled:      true,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the scraper name
func (s *HackerNewsScraper) Name() string {
	return "HackerNews"
}

// IsEnabled returns whether the scraper is enabled
func (s *HackerNewsScraper) IsEnabled() bool {
	return s.enabled
}

// Scrape performs HackerNews scraping
func (s *HackerNewsScraper) Scrape(ctx context.Context) ([]*entities.Notice, error) {
	s.logger.WithFields(map[string]interface{}{
		"current_id":     s.currentID,
		"max_jobs":       s.maxJobsToGet,
	}).Info("Starting HackerNews scraping")

	// Get the parent post
	parent, err := s.getParentPost(ctx)
	if err != nil {
		return nil, errors.WrapScraperError(err, "failed to get HackerNews parent post")
	}

	if len(parent.Kids) == 0 {
		s.logger.Warn("No child posts found in HackerNews parent")
		return []*entities.Notice{}, nil
	}

	// Sort kids by ID (newer posts have higher IDs)
	kidsCopy := make([]int, len(parent.Kids))
	copy(kidsCopy, parent.Kids)
	sort.Slice(kidsCopy, func(i, j int) bool {
		return kidsCopy[i] > kidsCopy[j]
	})

	// Limit the number of posts to process
	if s.maxJobsToGet > len(kidsCopy) {
		s.maxJobsToGet = len(kidsCopy)
	}
	kidsToProcess := kidsCopy[:s.maxJobsToGet]

	// Scrape stories concurrently
	stories := s.scrapeStories(ctx, kidsToProcess)

	// Convert stories to notices
	var notices []*entities.Notice
	for _, story := range stories {
		notice := s.convertStoryToNotice(story)
		if s.isValidNotice(notice) {
			notices = append(notices, notice)
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"total_stories": len(stories),
		"valid_notices": len(notices),
	}).Info("Completed HackerNews scraping")

	return notices, nil
}

// getParentPost fetches the parent HackerNews post
func (s *HackerNewsScraper) getParentPost(ctx context.Context) (*HackerNewsParent, error) {
	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json?print=pretty", s.currentID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parent HackerNewsParent
	if err := json.NewDecoder(resp.Body).Decode(&parent); err != nil {
		return nil, err
	}

	return &parent, nil
}

// scrapeStories fetches multiple stories concurrently
func (s *HackerNewsScraper) scrapeStories(ctx context.Context, storyIDs []int) []HackerNewsStory {
	var wg sync.WaitGroup
	storyCh := make(chan HackerNewsStory, len(storyIDs))

	for _, id := range storyIDs {
		wg.Add(1)
		go s.scrapeStory(ctx, id, &wg, storyCh)
	}

	wg.Wait()
	close(storyCh)

	var stories []HackerNewsStory
	for story := range storyCh {
		stories = append(stories, story)
	}

	return stories
}

// scrapeStory fetches a single story
func (s *HackerNewsScraper) scrapeStory(ctx context.Context, id int, wg *sync.WaitGroup, storyCh chan<- HackerNewsStory) {
	defer wg.Done()

	storyLogger := s.logger.WithFields(map[string]interface{}{
		"story_id": id,
	})

	// Check context cancellation
	select {
	case <-ctx.Done():
		storyLogger.Debug("Context cancelled, skipping story")
		return
	default:
	}

	url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json?print=pretty", id)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		storyLogger.WithError(err).Error("Failed to create request for story")
		return
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		storyLogger.WithError(err).Error("Failed to fetch story")
		return
	}
	defer resp.Body.Close()

	storyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		storyLogger.WithError(err).Error("Failed to read story response")
		return
	}

	var story HackerNewsStory
	if err := json.Unmarshal(storyBytes, &story); err != nil {
		storyLogger.WithError(err).Error("Failed to decode story JSON")
		return
	}

	story.Raw = string(storyBytes)
	storyCh <- story
}

// convertStoryToNotice converts a HackerNews story to a Notice entity
func (s *HackerNewsScraper) convertStoryToNotice(story HackerNewsStory) *entities.Notice {
	publishedTime := time.Unix(story.Time, 0).UTC()
	title, body := s.extractTitleAndBody(story.Text)

	return &entities.Notice{
		ID:            uuid.New().String(),
		Title:         title,
		Body:          body,
		URL:           fmt.Sprintf("https://news.ycombinator.com/item?id=%d", story.ID),
		AuthorName:    story.By,
		AuthorURL:     fmt.Sprintf("https://news.ycombinator.com/user?id=%s", story.By),
		SourceID:      "HackerNews",
		Raw:           story.Raw,
		Guid:          fmt.Sprint(story.ID),
		PublishedDate: &publishedTime,
	}
}

// extractTitleAndBody extracts title and body from HackerNews text
func (s *HackerNewsScraper) extractTitleAndBody(text string) (string, string) {
	title := ""
	body := text

	if idx := strings.Index(text, "<"); idx != -1 {
		title = text[:idx]
		body = text[idx:]
	} else {
		title = text
		if len(text) > 20 {
			title = text[:20]
			body = text[20:]
		}
	}

	// Remove unclosed brackets in the title
	stack := []rune{}
	newTitle := []rune{}

	for _, char := range title {
		if char == '(' || char == '[' || char == '{' {
			stack = append(stack, char)
		} else if char == ')' || char == ']' || char == '}' {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		} else {
			newTitle = append(newTitle, char)
		}
	}

	title = string(newTitle)

	// Decode HTML entities
	title = html.UnescapeString(title)
	body = html.UnescapeString(body)

	return title, body
}

// isValidNotice checks if a notice meets minimum quality criteria
func (s *HackerNewsScraper) isValidNotice(notice *entities.Notice) bool {
	return len(notice.Title) >= 10 || len(notice.Body) >= 10
}
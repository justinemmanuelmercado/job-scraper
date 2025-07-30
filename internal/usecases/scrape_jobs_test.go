package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/services"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/logging"
)

// MockNoticeRepository implements NoticeRepository for testing
type MockNoticeRepository struct {
	notices []entities.Notice
	count   int
}

func (m *MockNoticeRepository) Create(ctx context.Context, notices []*entities.Notice) error {
	for _, notice := range notices {
		m.notices = append(m.notices, *notice)
	}
	return nil
}

func (m *MockNoticeRepository) GetLatest(ctx context.Context, limit int) ([]*entities.Notice, error) {
	result := make([]*entities.Notice, 0, limit)
	for i := len(m.notices) - 1; i >= 0 && len(result) < limit; i-- {
		result = append(result, &m.notices[i])
	}
	return result, nil
}

func (m *MockNoticeRepository) GetCount(ctx context.Context) (int, error) {
	return len(m.notices), nil
}

func (m *MockNoticeRepository) GetLatestByDate(ctx context.Context, since time.Time) ([]*entities.Notice, error) {
	var result []*entities.Notice
	for i := range m.notices {
		if m.notices[i].CreatedAt.After(since) {
			result = append(result, &m.notices[i])
		}
	}
	return result, nil
}

func (m *MockNoticeRepository) GetBySource(ctx context.Context, sourceID string, limit int) ([]*entities.Notice, error) {
	var result []*entities.Notice
	count := 0
	for i := range m.notices {
		if m.notices[i].SourceID == sourceID && count < limit {
			result = append(result, &m.notices[i])
			count++
		}
	}
	return result, nil
}

// MockScrapingService implements ScrapingService for testing
type MockScrapingService struct {
	results []*entities.ScrapeResult
}

func (m *MockScrapingService) ScrapeAll(ctx context.Context) ([]*entities.ScrapeResult, error) {
	return m.results, nil
}

func (m *MockScrapingService) RegisterScraper(scraper services.Scraper) {}
func (m *MockScrapingService) GetScrapers() []services.Scraper        { return nil }

// MockNotificationService implements NotificationService for testing
type MockNotificationService struct {
	notifiedJobs    []*entities.Notice
	notifiedSummary *entities.Summary
}

func (m *MockNotificationService) NotifyAll(ctx context.Context, notices []*entities.Notice, summary *entities.Summary) error {
	m.notifiedJobs = notices
	m.notifiedSummary = summary
	return nil
}

func (m *MockNotificationService) RegisterNotifier(notifier services.Notifier) {}
func (m *MockNotificationService) GetNotifiers() []services.Notifier         { return nil }

func TestScrapeJobsUseCase_Execute(t *testing.T) {
	// Setup
	logger := logging.NewLogger("debug")
	mockRepo := &MockNoticeRepository{}
	mockScraping := &MockScrapingService{
		results: []*entities.ScrapeResult{
			{
				SourceName:   "TestSource",
				Notices:      []*entities.Notice{{ID: "1", Title: "Test Job", SourceID: "TestSource"}},
				ScrapedCount: 1,
				NewCount:     0,
			},
		},
	}
	mockNotify := &MockNotificationService{}

	useCase := NewScrapeJobsUseCase(mockRepo, mockScraping, mockNotify, logger)

	// Execute
	ctx := context.Background()
	err := useCase.Execute(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that notices were stored
	if len(mockRepo.notices) != 1 {
		t.Errorf("Expected 1 notice to be stored, got %d", len(mockRepo.notices))
	}

	// Check that notifications were sent
	if mockNotify.notifiedSummary == nil {
		t.Error("Expected summary notification to be sent")
	}

	if mockNotify.notifiedSummary.TotalScraped != 1 {
		t.Errorf("Expected TotalScraped to be 1, got %d", mockNotify.notifiedSummary.TotalScraped)
	}
}

func TestScrapeJobsUseCase_ExecuteWithNoResults(t *testing.T) {
	// Setup
	logger := logging.NewLogger("debug")
	mockRepo := &MockNoticeRepository{}
	mockScraping := &MockScrapingService{results: []*entities.ScrapeResult{}}
	mockNotify := &MockNotificationService{}

	useCase := NewScrapeJobsUseCase(mockRepo, mockScraping, mockNotify, logger)

	// Execute
	ctx := context.Background()
	err := useCase.Execute(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that no notices were stored
	if len(mockRepo.notices) != 0 {
		t.Errorf("Expected 0 notices to be stored, got %d", len(mockRepo.notices))
	}
}
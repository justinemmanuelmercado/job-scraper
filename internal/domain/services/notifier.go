package services

import (
	"context"

	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
)

// Notifier defines the interface for sending notifications
type Notifier interface {
	// Name returns the human-readable name of the notifier
	Name() string
	
	// SendJobs sends a list of job notices
	SendJobs(ctx context.Context, notices []*entities.Notice) error
	
	// SendSummary sends a summary of the scraping operation
	SendSummary(ctx context.Context, summary *entities.Summary) error
	
	// IsEnabled returns whether this notifier is currently enabled
	IsEnabled() bool
}

// NotificationService orchestrates multiple notifiers
type NotificationService interface {
	// NotifyAll sends notifications through all enabled notifiers
	NotifyAll(ctx context.Context, notices []*entities.Notice, summary *entities.Summary) error
	
	// RegisterNotifier adds a new notifier to the service
	RegisterNotifier(notifier Notifier)
	
	// GetNotifiers returns all registered notifiers
	GetNotifiers() []Notifier
}
package notifiers

import (
	"context"
	"sync"

	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/services"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/logging"
)

// NotificationServiceImpl implements the NotificationService interface
type NotificationServiceImpl struct {
	notifiers []services.Notifier
	logger    *logging.Logger
}

// NewNotificationService creates a new notification service
func NewNotificationService(logger *logging.Logger) services.NotificationService {
	return &NotificationServiceImpl{
		notifiers: make([]services.Notifier, 0),
		logger:    logger,
	}
}

// RegisterNotifier adds a new notifier to the service
func (s *NotificationServiceImpl) RegisterNotifier(notifier services.Notifier) {
	s.notifiers = append(s.notifiers, notifier)
	s.logger.WithFields(map[string]interface{}{
		"notifier_name": notifier.Name(),
		"enabled":       notifier.IsEnabled(),
	}).Info("Registered notifier")
}

// GetNotifiers returns all registered notifiers
func (s *NotificationServiceImpl) GetNotifiers() []services.Notifier {
	return s.notifiers
}

// NotifyAll sends notifications through all enabled notifiers
func (s *NotificationServiceImpl) NotifyAll(ctx context.Context, notices []*entities.Notice, summary *entities.Summary) error {
	enabledNotifiers := s.getEnabledNotifiers()
	
	if len(enabledNotifiers) == 0 {
		s.logger.Warn("No enabled notifiers found")
		return nil
	}

	if len(notices) == 0 {
		s.logger.Info("No notices to send, only sending summary")
		return s.sendSummaryToAll(ctx, enabledNotifiers, summary)
	}

	s.logger.WithFields(map[string]interface{}{
		"notifier_count": len(enabledNotifiers),
		"notice_count":   len(notices),
	}).Info("Starting notifications")

	var wg sync.WaitGroup
	errCh := make(chan error, len(enabledNotifiers)*2) // *2 for jobs + summary

	// Send notifications concurrently
	for _, notifier := range enabledNotifiers {
		wg.Add(2) // One for jobs, one for summary
		go s.sendJobs(ctx, notifier, notices, &wg, errCh)
		go s.sendSummary(ctx, notifier, summary, &wg, errCh)
	}

	wg.Wait()
	close(errCh)

	// Collect errors (but don't fail completely if some notifications fail)
	var notificationErrors []error
	for err := range errCh {
		if err != nil {
			notificationErrors = append(notificationErrors, err)
			s.logger.WithError(err).Warn("Notification error occurred")
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"error_count": len(notificationErrors),
	}).Info("Completed notifications")

	return nil
}

// getEnabledNotifiers returns only the enabled notifiers
func (s *NotificationServiceImpl) getEnabledNotifiers() []services.Notifier {
	var enabled []services.Notifier
	for _, notifier := range s.notifiers {
		if notifier.IsEnabled() {
			enabled = append(enabled, notifier)
		}
	}
	return enabled
}

// sendJobs sends job notices through a single notifier
func (s *NotificationServiceImpl) sendJobs(ctx context.Context, notifier services.Notifier, notices []*entities.Notice, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	notifierLogger := s.logger.WithFields(map[string]interface{}{
		"notifier_name": notifier.Name(),
		"notice_count":  len(notices),
	})

	notifierLogger.Debug("Sending job notices")

	if err := notifier.SendJobs(ctx, notices); err != nil {
		notifierLogger.WithError(err).Error("Failed to send job notices")
		errCh <- err
	} else {
		notifierLogger.Info("Successfully sent job notices")
		errCh <- nil
	}
}

// sendSummary sends summary through a single notifier
func (s *NotificationServiceImpl) sendSummary(ctx context.Context, notifier services.Notifier, summary *entities.Summary, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	notifierLogger := s.logger.WithFields(map[string]interface{}{
		"notifier_name": notifier.Name(),
	})

	notifierLogger.Debug("Sending summary")

	if err := notifier.SendSummary(ctx, summary); err != nil {
		notifierLogger.WithError(err).Error("Failed to send summary")
		errCh <- err
	} else {
		notifierLogger.Info("Successfully sent summary")
		errCh <- nil
	}
}

// sendSummaryToAll sends summary to all notifiers when there are no new notices
func (s *NotificationServiceImpl) sendSummaryToAll(ctx context.Context, notifiers []services.Notifier, summary *entities.Summary) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(notifiers))

	for _, notifier := range notifiers {
		wg.Add(1)
		go s.sendSummary(ctx, notifier, summary, &wg, errCh)
	}

	wg.Wait()
	close(errCh)

	// Collect errors
	var errors []error
	for err := range errCh {
		if err != nil {
			errors = append(errors, err)
		}
	}

	return nil // Don't fail even if some notifications fail
}
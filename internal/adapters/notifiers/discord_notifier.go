package notifiers

import (
	"context"
	"fmt"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/bwmarrin/discordgo"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/entities"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/services"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/config"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/errors"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/logging"
)

// DiscordNotifier implements the Notifier interface for Discord
type DiscordNotifier struct {
	session   *discordgo.Session
	config    config.DiscordConfig
	logger    *logging.Logger
	enabled   bool
	converter *md.Converter
}

// NewDiscordNotifier creates a new Discord notifier
func NewDiscordNotifier(cfg config.DiscordConfig, logger *logging.Logger) services.Notifier {
	if !cfg.Enabled || cfg.Token == "" {
		logger.Warn("Discord notifier is disabled or missing token")
		return &DiscordNotifier{enabled: false, logger: logger}
	}

	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		logger.WithError(err).Error("Failed to create Discord session")
		return &DiscordNotifier{enabled: false, logger: logger}
	}

	converter := md.NewConverter("", true, nil)

	return &DiscordNotifier{
		session:   session,
		config:    cfg,
		logger:    logger,
		enabled:   true,
		converter: converter,
	}
}

// Name returns the notifier name
func (n *DiscordNotifier) Name() string {
	return "Discord"
}

// IsEnabled returns whether the notifier is enabled
func (n *DiscordNotifier) IsEnabled() bool {
	return n.enabled
}

// SendJobs sends job notices to Discord
func (n *DiscordNotifier) SendJobs(ctx context.Context, notices []*entities.Notice) error {
	if !n.enabled {
		return errors.NewNotifierError("Discord notifier is disabled")
	}

	if len(notices) == 0 {
		n.logger.Info("No notices to send to Discord")
		return nil
	}

	n.logger.WithFields(map[string]interface{}{
		"notice_count": len(notices),
	}).Info("Sending job notices to Discord")

	for _, notice := range notices {
		if err := n.sendJobNotice(ctx, notice); err != nil {
			n.logger.WithError(err).WithFields(map[string]interface{}{
				"notice_id":    notice.ID,
				"notice_title": notice.Title,
			}).Error("Failed to send job notice to Discord")
			// Continue with other notices even if one fails
		}
	}

	return nil
}

// SendSummary sends a scraping summary to Discord
func (n *DiscordNotifier) SendSummary(ctx context.Context, summary *entities.Summary) error {
	if !n.enabled {
		return errors.NewNotifierError("Discord notifier is disabled")
	}

	n.logger.Info("Sending scraping summary to Discord")

	message := n.formatSummaryMessage(summary)
	
	_, err := n.session.ChannelMessageSend(n.config.HiringChannelID, message)
	if err != nil {
		return errors.WrapNotifierError(err, "failed to send summary to Discord")
	}

	n.logger.Info("Successfully sent scraping summary to Discord")
	return nil
}

// sendJobNotice sends a single job notice to Discord
func (n *DiscordNotifier) sendJobNotice(ctx context.Context, notice *entities.Notice) error {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return errors.WrapNotifierError(ctx.Err(), "context cancelled during Discord notification")
	default:
	}

	// Convert HTML to Markdown
	body, err := n.converter.ConvertString(notice.Body)
	if err != nil {
		n.logger.WithError(err).Warn("Failed to convert HTML to Markdown, using original body")
		body = notice.Body
	}

	// Add source information
	body = fmt.Sprintf("From %s \n %s", notice.SourceID, body)

	// Truncate content if too long
	const maxBodyLength = 500
	const maxTitleLength = 250

	if len(body) > maxBodyLength {
		body = body[:maxBodyLength] + "..."
	}

	title := notice.Title
	if len(title) > maxTitleLength {
		title = title[:maxTitleLength] + "..."
	}

	// Add site link
	body = fmt.Sprintf("%s\n[View on site](%s)", body, n.convertToSiteURL(notice.ID))

	embed := &discordgo.MessageEmbed{
		Title:       title,
		URL:         notice.URL,
		Description: body,
		Color:       0x00ff00, // Green color
		Timestamp:   notice.CreatedAt.Format(time.RFC3339),
	}

	_, err = n.session.ChannelMessageSendEmbed(n.config.HiringChannelID, embed)
	if err != nil {
		return errors.WrapNotifierError(err, "failed to send embed to Discord")
	}

	return nil
}

// formatSummaryMessage formats the scraping summary for Discord
func (n *DiscordNotifier) formatSummaryMessage(summary *entities.Summary) string {
	message := fmt.Sprintf(
		"Successfully run scraper at: %s\nTotal notices found: %d\nNew notices: %d\nRuntime: %v\n",
		summary.Timestamp.Format("January 2, 2006 15:04:05"),
		summary.TotalScraped,
		summary.TotalNew,
		summary.Duration,
	)

	// Add per-source breakdown
	if len(summary.Sources) > 0 {
		message += "\n**Source Breakdown:**\n"
		for _, source := range summary.Sources {
			status := "✅"
			if len(source.ErrorMessages) > 0 {
				status = "❌"
			}
			message += fmt.Sprintf("%s %s: %d notices\n", status, source.SourceName, source.ScrapedCount)
		}
	}

	message += "\nSite URL: https://workfindy.com/"

	return message
}

// convertToSiteURL converts a notice ID to a site URL
func (n *DiscordNotifier) convertToSiteURL(id string) string {
	return fmt.Sprintf("https://workfindy.com/%s", id)
}
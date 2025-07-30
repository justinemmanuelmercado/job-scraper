package container

import (
	"database/sql"
	"fmt"

	"github.com/justinemmanuelmercado/go-scraper/internal/adapters/notifiers"
	adapterRepos "github.com/justinemmanuelmercado/go-scraper/internal/adapters/repositories"
	"github.com/justinemmanuelmercado/go-scraper/internal/adapters/scrapers"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/repositories"
	"github.com/justinemmanuelmercado/go-scraper/internal/domain/services"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/config"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/database"
	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/logging"
	"github.com/justinemmanuelmercado/go-scraper/internal/usecases"
)

// Container holds all application dependencies
type Container struct {
	Config           *config.Config
	Logger           *logging.Logger
	DB               *sql.DB
	NoticeRepository repositories.NoticeRepository
	SourceRepository repositories.SourceRepository
	ScrapingService  services.ScrapingService
	NotifyService    services.NotificationService
	ScrapeJobsUC     *usecases.ScrapeJobsUseCase
}

// NewContainer creates and configures all application dependencies
func NewContainer() (*Container, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	logger := logging.NewLogger(cfg.App.LogLevel)
	appLogger := logger.WithComponent("container")

	// Initialize database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	appLogger.Info("Database connection established")

	// Initialize repositories
	noticeRepo := adapterRepos.NewPostgresNoticeRepository(db, logger.WithComponent("notice_repo"))
	sourceRepo := adapterRepos.NewPostgresSourceRepository(db, logger.WithComponent("source_repo"))

	// Initialize scraping service
	scrapingService := scrapers.NewScrapingService(logger.WithComponent("scraping_service"))

	// Register scrapers
	rssFeeds := []scrapers.RSSFeed{
		{URL: "https://weworkremotely.com/categories/remote-full-stack-programming-jobs.rss", SourceName: "WeWorkRemotely"},
		{URL: "https://weworkremotely.com/categories/remote-front-end-programming-jobs.rss", SourceName: "WeWorkRemotely"},
		{URL: "https://weworkremotely.com/categories/remote-back-end-programming-jobs.rss", SourceName: "WeWorkRemotely"},
		{URL: "https://remotive.com/remote-jobs/feed/software-dev", SourceName: "Remotive"},
		{URL: "https://jobicy.com/?feed=job_feed&job_categories=dev&job_types=full-time", SourceName: "JobIcy"},
		{URL: "https://www.fossjobs.net/rss/all/", SourceName: "FOSSJobs"},
		{URL: "https://remoteok.io/remote-jobs.rss", SourceName: "RemoteOK"},
	}

	rssScraper := scrapers.NewRSSFeedScraper(rssFeeds, logger.WithComponent("rss_scraper"))
	scrapingService.RegisterScraper(rssScraper)

	redditScraper := scrapers.NewRedditScraper(cfg.Reddit, logger.WithComponent("reddit_scraper"))
	scrapingService.RegisterScraper(redditScraper)

	hackerNewsScraper := scrapers.NewHackerNewsScraper(cfg.App.HackerNewsID, cfg.App.MaxJobsToGet, logger.WithComponent("hackernews_scraper"))
	scrapingService.RegisterScraper(hackerNewsScraper)

	// Initialize notification service
	notifyService := notifiers.NewNotificationService(logger.WithComponent("notification_service"))

	// Register notifiers
	if cfg.Discord.Enabled {
		discordNotifier := notifiers.NewDiscordNotifier(cfg.Discord, logger.WithComponent("discord_notifier"))
		notifyService.RegisterNotifier(discordNotifier)
	}

	// Initialize use cases
	scrapeJobsUC := usecases.NewScrapeJobsUseCase(
		noticeRepo,
		scrapingService,
		notifyService,
		logger.WithComponent("scrape_jobs_uc"),
	)

	appLogger.Info("All dependencies initialized successfully")

	return &Container{
		Config:           cfg,
		Logger:           logger,
		DB:               db,
		NoticeRepository: noticeRepo,
		SourceRepository: sourceRepo,
		ScrapingService:  scrapingService,
		NotifyService:    notifyService,
		ScrapeJobsUC:     scrapeJobsUC,
	}, nil
}

// Close cleans up all resources
func (c *Container) Close() error {
	c.Logger.Info("Shutting down application...")
	
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			c.Logger.WithError(err).Error("Failed to close database connection")
			return err
		}
	}
	
	c.Logger.Info("Application shutdown complete")
	return nil
}
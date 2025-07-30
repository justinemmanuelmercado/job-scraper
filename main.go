package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/justinemmanuelmercado/go-scraper/internal/infrastructure/container"
)

func main() {
	// Parse command line flags
	genMarkdown := flag.Bool("markdown", false, "Generate Markdown file for latest notices")
	flag.Parse()

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		log.Println("Received shutdown signal, initiating graceful shutdown...")
		cancel()
	}()

	// Initialize dependency container
	container, err := container.NewContainer()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer func() {
		if closeErr := container.Close(); closeErr != nil {
			container.Logger.WithError(closeErr).Error("Error during application shutdown")
		}
	}()

	logger := container.Logger.WithComponent("main")
	logger.Info("Application started successfully")

	// Add timeout to context for scraping operations
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, container.Config.App.GracefulTimeout)
	defer timeoutCancel()

	// Execute the appropriate workflow
	if *genMarkdown {
		logger.Info("Starting markdown generation workflow")
		if err := container.ScrapeJobsUC.ExecuteWithMarkdownGeneration(timeoutCtx); err != nil {
			logger.WithError(err).Error("Markdown generation workflow failed")
			os.Exit(1)
		}
		logger.Info("Markdown generation workflow completed successfully")
	} else {
		logger.Info("Starting job scraping workflow")
		if err := container.ScrapeJobsUC.Execute(timeoutCtx); err != nil {
			logger.WithError(err).Error("Job scraping workflow failed")
			os.Exit(1)
		}
		logger.Info("Job scraping workflow completed successfully")
	}
}
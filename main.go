package main

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/justinemmanuelmercado/go-scraper/pkg/discord"
	"github.com/justinemmanuelmercado/go-scraper/pkg/models"
	"github.com/justinemmanuelmercado/go-scraper/pkg/reddit"
	"github.com/justinemmanuelmercado/go-scraper/pkg/rss_feed"
	"github.com/justinemmanuelmercado/go-scraper/pkg/store"
	"log"
	"os"
)

func loadEnvFile() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading environment variables %w", err)
	}
	return nil
}

func insertNotices(newNotices []*models.Notice, db *pgx.Conn) error {
	noticeStore := store.InitNotice(db)
	return noticeStore.CreateNotices(newNotices)
}

func setUpDatabase() (*pgx.Conn, error) {
	dbUrl := os.Getenv("DATABASE_URL")
	db, err := store.OpenDB(dbUrl)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return db, nil
}

func getRssFeedNotices(source *store.Source) ([]*models.Notice, error) {
	newNotices, err := rss_feed.GetAllNotices(source)
	if err != nil {
		return nil, fmt.Errorf("error getting Notices from RSS feeds: %w", err)
	}

	return newNotices, nil
}

func getRedditNotices(source *store.Source) ([]*models.Notice, error) {
	return reddit.GetNoticesFromPosts(source)
}

func main() {
	err := loadEnvFile()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v\n", err)
	}

	db, err := setUpDatabase()
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}

	source := store.InitSource(db)

	rssFeedNotices, err := getRssFeedNotices(source)
	if err != nil {
		log.Fatalf("error getting notices from rss feeds: %v\n", err)
	}

	redditNotices, err := getRedditNotices(source)
	if err != nil {
		log.Fatalf("error getting notices from reddit: %v\n", err)
	}

	allNotices := append(rssFeedNotices, redditNotices...)
	log.Printf("Trying to insert %d notices \n", len(allNotices))

	err = insertNotices(allNotices, db)

	if err != nil {
		log.Fatalf("Error creating notices: %v\n", err)
	}

	err = discord.SendNotification(fmt.Sprintf("Succesfully run script with %d notices matched", len(allNotices)))
	if err != nil {
		log.Println("Discord notification not sent")
	}
	log.Println("Script run successfully")

}

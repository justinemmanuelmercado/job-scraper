package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/justinemmanuelmercado/go-scraper/pkg/discord"
	"github.com/justinemmanuelmercado/go-scraper/pkg/models"
	"github.com/justinemmanuelmercado/go-scraper/pkg/reddit"
	"github.com/justinemmanuelmercado/go-scraper/pkg/rss_feed"
	"github.com/justinemmanuelmercado/go-scraper/pkg/store"
)

func loadEnvFile() {
	err := godotenv.Load()
	if err != nil {
		log.Println("error loading environment variables %w", err)
	}
}

func setUpDatabase() (*pgx.Conn, error) {
	dbPw := os.Getenv("POSTGRES_PASSWORD")
	dbUser := os.Getenv("POSTGRES_USER")
	dbName := os.Getenv("POSTGRES_DB")
	dbUrl := fmt.Sprintf("postgresql://%s:%s@localhost:5432/%s?sslmode=disable", dbUser, dbPw, dbName)
	db, err := store.OpenDB(dbUrl)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
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

func main() {
	startTime := time.Now()
	loadEnvFile()

	db, err := setUpDatabase()
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}

	source := store.InitSource(db)

	rssFeedNotices, err := getRssFeedNotices(source)
	if err != nil {
		log.Fatalf("error getting notices from rss feeds: %v\n", err)
	}

	redditNotices, err := reddit.GetNoticesFromPosts(source)
	if err != nil {
		log.Fatalf("error getting notices from reddit: %v\n", err)
	}

	allNotices := append(rssFeedNotices, redditNotices...)
	log.Printf("Trying to insert %d notices \n", len(allNotices))

	noticeStore := store.InitNotice(db)
	oldNoticeCount := noticeStore.GetCount()
	err = noticeStore.CreateNotices(allNotices)

	if err != nil {
		log.Fatalf("Error creating notices: %v\n", err)
	}

	newNoticeCount := noticeStore.GetCount()

	err = discord.SendNotification(len(allNotices), newNoticeCount-oldNoticeCount, time.Since(startTime))
	if err != nil {
		log.Println("Discord notification not sent")
	}
	log.Println("Script run successfully")

}

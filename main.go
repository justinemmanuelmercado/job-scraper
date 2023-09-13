package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	errorHandler "github.com/justinemmanuelmercado/go-scraper/pkg"
	"github.com/justinemmanuelmercado/go-scraper/pkg/discord"
	"github.com/justinemmanuelmercado/go-scraper/pkg/hackernews"
	"github.com/justinemmanuelmercado/go-scraper/pkg/models"
	"github.com/justinemmanuelmercado/go-scraper/pkg/reddit"
	"github.com/justinemmanuelmercado/go-scraper/pkg/rss_feed"
	"github.com/justinemmanuelmercado/go-scraper/pkg/store"
)

func init() {
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

func getRssFeedNotices() ([]*models.Notice, error) {
	newNotices, err := rss_feed.GetAllNotices()
	if err != nil {
		return nil, fmt.Errorf("error getting Notices from RSS feeds: %w", err)
	}

	return newNotices, nil
}

func getHackerNews() []*models.Notice {
	newNotices := hackernews.ScrapeCurrentWhoIsHiringPosts()
	return newNotices
}

func scrape() {
	startTime := time.Now()

	db, err := setUpDatabase()
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}

	rssFeedNotices, err := getRssFeedNotices()
	errorHandler.HandleErrorWithSection(err, "Failed to get notices from rss feeds", "RSS Feeds")

	redditNotices, err := reddit.GetNoticesFromPosts()
	errorHandler.HandleErrorWithSection(err, "Failed to get notices from reddit", "Reddit")

	hnNotices := getHackerNews()

	allNotices := append(rssFeedNotices, redditNotices...)
	allNotices = append(allNotices, hnNotices...)
	log.Printf("Trying to insert %d notices \n", len(allNotices))

	noticeStore := store.InitNotice(db)
	oldNoticeCount := noticeStore.GetCount()
	err = noticeStore.CreateNotices(allNotices)
	if err != nil {
		log.Fatalf("Error inserting notices: %v\n", err)
	}

	newNoticeCount := noticeStore.GetCount()
	noticesInserted := newNoticeCount - oldNoticeCount

	if noticesInserted == 0 {
		log.Println("No new notices inserted")
		return
	}

	latestPosts := noticeStore.GetLatest(noticesInserted)

	bot, dscErr := discord.InitDiscordClient()

	if dscErr == nil {
		discord.SendList(bot, latestPosts)
		err = discord.SendSuccessNotificatioin(bot, len(allNotices), noticesInserted, time.Since(startTime))
		errorHandler.HandleErrorWithSection(err, "Failed to send success notification", "Discord")
	} else {
		log.Println("Discord client not initialized")
	}

	log.Println("Script run successfully")

}

func generateMarkdown() {
	// Initialize DB connection and NoticeStore
	db, err := setUpDatabase()
	if err != nil {
		fmt.Printf("error connecting to database: %v", err)
		return
	}
	store := store.InitNotice(db)

	// Fetch the latest notices
	notices, err := store.GetLatestNotices()
	if err != nil {
		fmt.Println("Error fetching notices:", err)
		return
	}

	// Create the folder if it doesn't exist
	if _, err := os.Stat("latest_notices"); os.IsNotExist(err) {
		os.Mkdir("latest_notices", 0755)
	}

	// Format the filename with the current date
	filename := fmt.Sprintf("latest_notices/%s_latest_notices.md", time.Now().Format("2006-01-02"))

	// Open a new markdown file
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating markdown file:", err)
		return
	}
	defer f.Close()

	// Write the header
	f.WriteString("# Latest Notices\n")
	f.WriteString(time.Now().Format("2006-01-02") + "\n\n")

	// Loop through the notices and write them to the file
	for _, notice := range notices {
		f.WriteString(fmt.Sprintf("## %s\n", notice.Title))
		f.WriteString(fmt.Sprintf("**Author**: %s\n", notice.AuthorName))
		f.WriteString(fmt.Sprintf("**Read more**: [Here](https://workfindy.com/%s)\n", notice.ID))
		f.WriteString("---\n")
	}
}

func main() {
	// Define a flag
	genMarkdown := flag.Bool("markdown", false, "Generate Markdown file for latest notices")
	flag.Parse()

	if *genMarkdown {
		generateMarkdown()
	} else {
		scrape()
	}
}

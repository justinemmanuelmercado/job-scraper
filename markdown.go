package main

import (
	"fmt"
	"html"
	"os"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/justinemmanuelmercado/go-scraper/pkg/store"
)

func printToHTML(text string) string {
	return html.UnescapeString(text)
}

func printToHTMLAndTruncate(text string, length int) string {
	decodedText := printToHTML(text)
	// Use regex to remove HTML tags
	re := regexp2.MustCompile("<[^>]*>?", regexp2.None)
	strippedText, _ := re.Replace(decodedText, "", 0, -1)

	// Unescape the string again because of weird encoded characters
	finalText := html.UnescapeString(strippedText)

	// Truncate the string
	if len(finalText) > length {
		return finalText[:length] + "..."
	}
	return finalText
}

func GenerateMarkdown() {
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

		f.WriteString(fmt.Sprintf("## %s\n\n", printToHTMLAndTruncate(notice.Title, 100)))
		f.WriteString(fmt.Sprintf("**From**: %s\n\n", notice.SourceID))
		f.WriteString(fmt.Sprintf("%s\n\n", printToHTMLAndTruncate(notice.Body, 500)))
		f.WriteString(fmt.Sprintf("**Read more**: [Here](https://workfindy.com/%s)\n\n", html.EscapeString(notice.ID)))
		f.WriteString("---\n\n")
	}
}

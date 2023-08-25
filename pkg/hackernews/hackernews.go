package hackernews

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	errorHandler "github.com/justinemmanuelmercado/go-scraper/pkg"
	"github.com/justinemmanuelmercado/go-scraper/pkg/models"
)

type Story struct {
	Time int64  `json:"time"`
	Text string `json:"text"`
	Id   int    `json:"id"`
	By   string `json:"by"`
	Raw  string `json:"raw"`
}

type Parent struct {
	Kids []int
	Time int64
}

var hnSourceName = "HackerNews"
var currentId = 36956867 // Automate getting this
var toGet = 20           // Only need 20 I think

func extractTitle(text string) (string, string) {
	title := ""
	body := text

	if idx := strings.Index(text, "<"); idx != -1 {
		title = text[:idx]
		body = text[idx:]
	} else {
		title = text
		if len(text) > 20 {
			title = text[:20]
			body = text[20:]
		}
	}

	return title, body
}

func getStory(id int, wg *sync.WaitGroup, ch chan<- Story) {
	defer wg.Done()
	storyResp, err := http.Get(fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json?print=pretty", id))
	if err != nil {
		errorHandler.HandleErrorWithSection(err, fmt.Sprintf("Unable to get story with ID of %d\n", id), "HackerNews")
		return
	}
	defer storyResp.Body.Close()

	storyBytes, err := io.ReadAll(storyResp.Body)
	if err != nil {
		errorHandler.HandleErrorWithSection(err, fmt.Sprintf("Unable to convert response with ID of %d to string\n", id), "HackerNews")
		return
	}

	var story Story
	err = json.Unmarshal(storyBytes, &story)
	if err != nil {
		errorHandler.HandleErrorWithSection(err, fmt.Sprintf("Unable to decode response with ID of %d\n", id), "HackerNews")
		return
	}

	story.Raw = string(storyBytes)
	ch <- story
}

func ScrapeCurrentWhoIsHiringPosts() []*models.Notice {
	var notices []*models.Notice
	var parent Parent

	resp, err := http.Get(fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json?print=pretty", currentId))
	errorHandler.HandleErrorWithSection(err, "Unable to get current who is hiring post", "HackerNews")

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&parent); err != nil {
		errorHandler.HandleErrorWithSection(err, "Unable to decode response for current who is hiring post", "HackerNews")
		return nil
	}

	// Sort by ID because I think newer posts are have higher IDs
	kidsCopy := append([]int(nil), parent.Kids...)
	sort.Slice(kidsCopy, func(i, j int) bool { return kidsCopy[i] > kidsCopy[j] })
	if toGet > len(kidsCopy) {
		toGet = len(kidsCopy)
	}
	storyChannel := make(chan Story, len(kidsCopy[:toGet]))

	var wg sync.WaitGroup
	wg.Add(len(kidsCopy[:toGet]))

	for _, id := range kidsCopy[:toGet] {
		go getStory(id, &wg, storyChannel)
	}

	wg.Wait()
	close(storyChannel)

	for story := range storyChannel {
		s := StoryToNotice(story)
		notices = append(notices, &s)
	}

	fmt.Printf("Fetched %d items from HackerNews\n", len(notices))

	return notices
}

func urlify(id int) string {
	return fmt.Sprintf("https://news.ycombinator.com/item?id=%d", id)
}

func authorUrlify(authorName string) string {
	return fmt.Sprintf("https://news.ycombinator.com/user?id=%s", authorName)
}

func StoryToNotice(story Story) models.Notice {
	t := time.Unix(story.Time, 0).UTC()
	title, body := extractTitle(story.Text)
	return models.Notice{
		ID:            uuid.New().String(),
		Title:         title,
		Body:          body,
		URL:           urlify(story.Id),
		AuthorName:    story.By,
		AuthorURL:     authorUrlify(story.By),
		ImageURL:      nil,
		SourceID:      hnSourceName,
		Raw:           story.Raw,
		Guid:          fmt.Sprint(story.Id),
		PublishedDate: &t,
	}
}

package rss_feed

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/justinemmanuelmercado/go-scraper/pkg/models"
	"github.com/mmcdole/gofeed"
)

type RssFeed struct {
	url        string
	sourceName string
}

var RssFeedPairs = []RssFeed{
	{"https://weworkremotely.com/categories/remote-full-stack-programming-jobs.rss", "WeWorkRemotely"},
	{"https://weworkremotely.com/categories/remote-front-end-programming-jobs.rss", "WeWorkRemotely"},
	{"https://weworkremotely.com/categories/remote-back-end-programming-jobs.rss", "WeWorkRemotely"},
	{"https://remotive.com/remote-jobs/feed/software-dev", "Remotive"},
	{"https://jobicy.com/?feed=job_feed&job_categories=dev&job_types=full-time", "JobIcy"},
	{"https://www.fossjobs.net/rss/all/", "FOSSJobs"},
	{"https://remoteok.io/remote-jobs.rss", "RemoteOK"},
}

func (rf *RssFeed) FetchItems() ([]*gofeed.Item, error) {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(rf.url)
	if err != nil {
		return nil, err
	}

	return feed.Items, nil
}

func NoticesFromFeedItems(items []*gofeed.Item, sourceId string) []*models.Notice {
	notices := make([]*models.Notice, len(items))

	for i, item := range items {
		jsonData, err := json.Marshal(item)
		if err != nil {
			jsonData = []byte{}
		}

		newNotice := &models.Notice{
			ID:            uuid.New().String(),
			Title:         item.Title,
			Body:          item.Description,
			URL:           item.Link,
			SourceID:      sourceId,
			Raw:           string(jsonData),
			Guid:          item.GUID,
			PublishedDate: item.PublishedParsed,
		}

		if item.Image != nil {
			newNotice.ImageURL = &item.Image.URL
		}

		if len(item.Authors) > 0 {
			newNotice.AuthorName = item.Authors[0].Name
			newNotice.AuthorURL = item.Authors[0].Email
		}

		notices[i] = newNotice
	}

	return notices
}

func GetAllNotices() ([]*models.Notice, error) {

	var wg sync.WaitGroup
	noticesCh := make(chan []*models.Notice, len(RssFeedPairs))
	errCh := make(chan error, len(RssFeedPairs))

	handleFeed := func(feed RssFeed) {
		defer wg.Done()

		items, err := feed.FetchItems()
		if err != nil {
			errCh <- fmt.Errorf("failed to fetch %s: %w", feed.sourceName, err)
			return
		}
		fmt.Printf("Fetched %d items from %s\n", len(items), feed.sourceName)
		noticesCh <- NoticesFromFeedItems(items, feed.sourceName)
	}

	wg.Add(len(RssFeedPairs))

	for _, rssFeedPair := range RssFeedPairs {
		go handleFeed(rssFeedPair)
	}

	wg.Wait()

	close(noticesCh)
	close(errCh)

	if len(errCh) > 0 {
		// Don't stop the process if one feed fails
		log.Printf("error fetching feeds: %v", <-errCh)
	}

	var allNotices []*models.Notice
	for notices := range noticesCh {
		allNotices = append(allNotices, notices...)
	}

	return allNotices, nil
}

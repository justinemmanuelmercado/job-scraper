package rss_feed

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/justinemmanuelmercado/go-scraper/pkg/models"
	"github.com/justinemmanuelmercado/go-scraper/pkg/store"
	"github.com/mmcdole/gofeed"
	"sync"
)

type RssFeed struct {
	url        string
	sourceName string
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

func GetAllNotices(source *store.Source) ([]*models.Notice, error) {
	var feedFuncs = []func() *RssFeed{
		GetWwrFeed,
		GetRemotiveFeed,
		GetJobIcyFeed,
	}

	var wg sync.WaitGroup
	noticesCh := make(chan []*models.Notice, len(feedFuncs))
	errCh := make(chan error, len(feedFuncs))

	handleFeed := func(feedFunc func() *RssFeed) {
		defer wg.Done()

		feed := feedFunc()
		items, err := feed.FetchItems()
		if err != nil {
			errCh <- fmt.Errorf("failed to fetch %s: %w", feed.sourceName, err)
			return
		}

		sourceItem, err := source.GetSourceByName(feed.sourceName)
		if err != nil {
			errCh <- fmt.Errorf("failed to fetch %s source id: %w", feed.sourceName, err)
			return
		}
		noticesCh <- NoticesFromFeedItems(items, sourceItem.ID)
	}

	wg.Add(len(feedFuncs))

	for _, feedFunc := range feedFuncs {
		go handleFeed(feedFunc)
	}

	wg.Wait()

	close(noticesCh)
	close(errCh)

	if len(errCh) > 0 {
		return nil, <-errCh
	}

	var allNotices []*models.Notice
	for notices := range noticesCh {
		allNotices = append(allNotices, notices...)
	}

	return allNotices, nil
}

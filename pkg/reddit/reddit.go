package reddit

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/justinemmanuelmercado/go-scraper/pkg/models"
	"github.com/thecsw/mira"
)

type Handler struct {
	r          RedditClient
	subreddits []string
}

type RedditClient interface {
	GetSubredditPosts(sr string, sort string, duration string, limit int) ([]mira.PostListingChild, error)
}

func InitRedditHandler(r RedditClient, subreddits []string) (*Handler, error) {
	return &Handler{
		r:          r,
		subreddits: subreddits,
	}, nil

}

func GetRedditHandler() (*Handler, error) {
	var subreddits = []string{"forhire", "remotejs"}
	r, err := mira.Init(mira.Credentials{
		ClientId:     os.Getenv("REDDIT_ID"),
		ClientSecret: os.Getenv("REDDIT_SECRET"),
		Username:     os.Getenv("REDDIT_USERNAME"),
		Password:     os.Getenv("REDDIT_PASSWORD"),
		UserAgent:    `SALPHBot`,
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to reddit client %w", err)
	}

	return InitRedditHandler(r, subreddits)

}

func (h *Handler) GetRedditPosts() ([]mira.PostListingChild, error) {
	var wg sync.WaitGroup
	postsCh := make(chan []mira.PostListingChild, len(h.subreddits))
	errCh := make(chan error, len(h.subreddits))

	handleSubreddit := func(sr string) {
		defer wg.Done()
		posts, err := h.r.GetSubredditPosts(sr, "new", "week", 20)
		if err != nil {
			errCh <- fmt.Errorf("failed to fetch from #{subreddit}: #{err}")
			return
		}

		posts = FilterHiringPosts(posts)
		postsCh <- posts
	}

	wg.Add(len(h.subreddits))

	for _, sr := range h.subreddits {
		go handleSubreddit(sr)
	}

	wg.Wait()

	close(postsCh)
	close(errCh)

	if len(errCh) > 0 {
		return nil, <-errCh
	}

	var allPosts []mira.PostListingChild
	for posts := range postsCh {
		allPosts = append(allPosts, posts...)
	}

	return allPosts, nil
}

const redditSourceName = "Reddit"

func GetNoticesFromPosts() ([]*models.Notice, error) {
	handler, err := GetRedditHandler()
	if err != nil {
		return nil, fmt.Errorf("failed to load handler for reddit: %w", err)
	}

	posts, err := handler.GetRedditPosts()
	if err != nil {
		return nil, fmt.Errorf("failed to load posts from reddit: %w", err)
	}

	notices := make([]*models.Notice, len(posts))

	for i, post := range posts {
		jsonData, err := json.Marshal(post)
		if err != nil {
			jsonData = []byte{}
		}
		t := time.Unix(int64(post.GetTimeCreated()), 0).UTC()
		newNotice := &models.Notice{
			ID:            uuid.New().String(),
			Title:         post.Data.Title,
			Body:          html.UnescapeString(post.Data.SelftextHtml),
			URL:           fmt.Sprintf(`https://www.reddit.com/%s`, post.GetPermalink()),
			AuthorName:    post.GetAuthor(),
			AuthorURL:     fmt.Sprintf(`https://www.reddit.com/user/%s`, post.GetAuthor()),
			ImageURL:      nil,
			SourceID:      redditSourceName,
			Raw:           string(jsonData),
			Guid:          post.Data.Id,
			PublishedDate: &t,
		}

		if len(post.Data.Preview.Images) > 0 {
			newNotice.ImageURL = &post.Data.Preview.Images[0].Source.Url
		}

		notices[i] = newNotice
	}
	fmt.Printf("Fetched %d items from %s \n", len(notices), redditSourceName)
	return notices, nil
}

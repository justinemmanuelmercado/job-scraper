package reddit

import (
	"errors"
	"github.com/thecsw/mira"
	"testing"
)

type mockRedditClient struct {
	posts []mira.PostListingChild
	err   error
}

func (m *mockRedditClient) GetSubredditPosts(sr string, sort string, duration string, limit int) ([]mira.PostListingChild, error) {
	return m.posts, m.err
}

func TestGetRedditPosts(t *testing.T) {
	testCases := []struct {
		name     string
		posts    []mira.PostListingChild
		err      error
		expected int
	}{
		{
			name:     "Successfully retrieve posts from multiple subreddits",
			posts:    []mira.PostListingChild{{Data: mira.PostListingChildData{Title: "post1"}}, {Data: mira.PostListingChildData{Title: "post2"}}},
			err:      nil,
			expected: 4,
		},
		{
			name:     "Error retrieving posts",
			posts:    nil,
			err:      errors.New("error retrieving posts"),
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mockRedditClient{
				posts: tc.posts,
				err:   tc.err,
			}

			handler, _ := InitRedditHandler(mockClient, []string{"subreddit1", "subreddit2"})
			posts, err := handler.GetRedditPosts()

			if err != nil && tc.err == nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil && tc.err != nil {
				t.Errorf("Expected error but got none")
			}

			if len(posts) != tc.expected {
				t.Errorf("Expected %d posts, got %d", tc.expected, len(posts))
			}
		})
	}
}

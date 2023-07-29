package reddit

import (
	"github.com/thecsw/mira"
	"strings"
)

func FilterHiringPosts(posts []mira.PostListingChild) []mira.PostListingChild {
	var hiringPosts []mira.PostListingChild
	for _, post := range posts {
		if strings.Contains(strings.ToUpper(post.Data.Title), "HIRING") {
			hiringPosts = append(hiringPosts, post)
		}
	}

	return hiringPosts
}

package rss_feed

const wwrRssFeedUrl = "https://weworkremotely.com/categories/remote-programming-jobs.rss"
const wwrSourceName = "WeWorkRemotely"

func GetWwrFeed() *RssFeed {
	return &RssFeed{wwrRssFeedUrl, wwrSourceName}
}

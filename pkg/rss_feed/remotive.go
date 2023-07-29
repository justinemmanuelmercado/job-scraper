package rss_feed

const remotiveUrl = "https://remotive.com/remote-jobs/feed/software-dev"
const remotiveSourceName = "Remotive"

func GetRemotiveFeed() *RssFeed {
	return &RssFeed{remotiveUrl, remotiveSourceName}
}

package rss_feed

const jobicyUrl = "https://jobicy.com/?feed=job_feed&job_categories=dev"
const jobicyName = "JobIcy"

func GetJobIcyFeed() *RssFeed {
	return &RssFeed{jobicyUrl, jobicyName}
}

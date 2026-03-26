package api

import "github.com/eldius/document-feeder/internal/model"

type AddFeedRequest struct {
	Feeds []string `json:"feeds"`
}

type FeedListResponse struct {
	Feeds []FeedSummary `json:"feeds"`
}

type FeedSummary struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Error string `json:"error,omitempty"`
}

func ToFeedSummary(feed *model.Feed) *FeedSummary {
	if feed == nil {
		return nil
	}
	return &FeedSummary{
		Title: feed.Title,
		URL:   feed.Link,
	}
}

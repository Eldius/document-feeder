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

type SearchRequest struct {
	Query string `json:"query"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

type SearchResult struct {
	FeedTitle  string  `json:"feed_title"`
	Article    Article `json:"article"`
	Similarity float32 `json:"similarity"`
}

type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Link        string `json:"link"`
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

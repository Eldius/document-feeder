package adapter

import (
	"context"
	"fmt"
	"github.com/eldius/document-feed-embedder/internal/feed"
	"github.com/eldius/document-feed-embedder/internal/model"
	"github.com/eldius/document-feed-embedder/internal/persistence/chromem"
	"github.com/eldius/document-feed-embedder/internal/persistence/storm"
)

type FeedAdapter struct {
	r    *storm.Repository
	docs *chromem.DocumentEmbedder
	p    feed.Parser
}

func NewFeedAdapter(r *storm.Repository, p feed.Parser, docs *chromem.DocumentEmbedder) *FeedAdapter {
	return &FeedAdapter{r: r, p: p, docs: docs}
}

func NewDefaultAdapter() (*FeedAdapter, error) {
	r := storm.NewRepository()
	p := feed.NewParser()
	d, err := chromem.NewDocumentEmbedder()
	if err != nil {
		return nil, fmt.Errorf("creating document embedder: %w", err)
	}
	return NewFeedAdapter(r, p, d), nil
}

func (a *FeedAdapter) Parse(ctx context.Context, feedURL string) (*model.Feed, error) {
	fmt.Println("parsing feed:", feedURL)
	f, err := a.p.Parse(ctx, feedURL)
	if err != nil {
		return nil, fmt.Errorf("parsing feed: %w", err)
	}

	if err := a.r.Persist(ctx, f); err != nil {
		return nil, fmt.Errorf("persisting feed: %w", err)
	}

	if err := a.docs.Save(ctx, f); err != nil {
		return nil, fmt.Errorf("saving article: %w", err)
	}

	return f, nil
}

func (a *FeedAdapter) Refresh(ctx context.Context) error {
	feeds, err := a.All(ctx)
	if err != nil {
		return fmt.Errorf("getting all feeds: %w", err)
	}

	for _, f := range feeds {
		nf, err := a.Parse(ctx, f.FeedLink)
		if err != nil {
			return fmt.Errorf("parsing feed: %w", err)
		}
		f.AddItems(nf.Items...)
		if err := a.r.Persist(ctx, f); err != nil {
			return fmt.Errorf("persisting feed: %w", err)
		}

		if err := a.docs.Save(ctx, f); err != nil {
			return fmt.Errorf("saving article: %w", err)
		}
	}
	return nil
}

func (a *FeedAdapter) All(ctx context.Context) ([]*model.Feed, error) {
	return a.r.All(ctx)
}

func (a *FeedAdapter) Search(ctx context.Context, term string) ([]*model.Article, error) {
	return a.docs.Search(ctx, term)
}

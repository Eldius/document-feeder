package adapter

import (
	"context"
	"embed"
	"fmt"
	"github.com/eldius/document-feed-embedder/internal/client"
	"github.com/eldius/document-feed-embedder/internal/feed"
	"github.com/eldius/document-feed-embedder/internal/model"
	"github.com/eldius/document-feed-embedder/internal/persistence/chromem"
	"github.com/eldius/document-feed-embedder/internal/persistence/storm"
	"slices"
	"strings"
	"text/template"
)

var (
	//go:embed templates/**
	templates embed.FS
)

type FeedAdapter struct {
	r      *storm.Repository
	docs   *chromem.DocumentVectorizer
	p      feed.Parser
	tmpl   *template.Template
	ollama *client.OllamaClient
}

func NewFeedAdapter(r *storm.Repository, p feed.Parser, docs *chromem.DocumentVectorizer, ollama *client.OllamaClient, tmpl *template.Template) *FeedAdapter {
	return &FeedAdapter{
		r:      r,
		p:      p,
		docs:   docs,
		tmpl:   tmpl,
		ollama: ollama,
	}
}

func NewDefaultAdapter() (*FeedAdapter, error) {
	r := storm.NewRepository()
	p := feed.NewParser()
	d, err := chromem.NewDefaultDocumentVectorizer()
	if err != nil {
		return nil, fmt.Errorf("creating document embedder: %w", err)
	}
	tmpl, err := template.ParseFS(templates, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}
	ollama := client.NewOllamaClient()
	return NewFeedAdapter(r, p, d, ollama, tmpl), nil
}

func (a *FeedAdapter) Parse(ctx context.Context, feedURL string) (*model.Feed, error) {
	fmt.Println("parsing feed:", feedURL)
	f, err := a.p.Parse(ctx, feedURL)
	if err != nil {
		return nil, fmt.Errorf("parsing feed: %w", err)
	}

	f.FeedLink = feedURL

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

func (a *FeedAdapter) Search(ctx context.Context, term string, maxResults int) ([]*model.SearchResult, error) {
	docs, err := a.docs.Search(ctx, term, maxResults)
	if err != nil {
		return nil, fmt.Errorf("searching documents: %w", err)
	}

	var res []*model.SearchResult
	for _, d := range docs {
		doc, err := a.r.ArticleByLink(ctx, d.Metadata["feed"], d.ID)
		if err != nil {
			return nil, fmt.Errorf("getting article by link: %w", err)
		}
		res = append(res, &model.SearchResult{
			FeedTitle: d.Metadata["feed"],
			Article: model.Article{
				Title:           doc.Title,
				Description:     doc.Description,
				Content:         doc.Content,
				Link:            doc.Link,
				Published:       doc.Published,
				PublishedParsed: doc.PublishedParsed,
				Authors:         doc.Authors,
			},
			Similarity:       d.Similarity,
			SanitizedContent: d.Content,
			Embeddings:       d.Embedding,
		})
	}

	return slices.SortedFunc(slices.Values(res), func(e *model.SearchResult, e2 *model.SearchResult) int {
		return int(e2.Similarity) - int(e.Similarity)
	}), nil
}

type promptTemplateData struct {
	Question string
	Results  []*model.SearchResult
}

func (a *FeedAdapter) AskAQuestion(ctx context.Context, question string) (string, error) {
	docs, err := a.Search(ctx, question, 10)
	if err != nil {
		return "", fmt.Errorf("searching documents: %w", err)
	}
	data := promptTemplateData{Question: question, Results: docs}

	var b strings.Builder
	if err := a.tmpl.ExecuteTemplate(&b, "prompt.tmpl", data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return a.ollama.GenerationFunc(
		ctx,
		b.String(),
		false,
	)
}

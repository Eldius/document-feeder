package feed

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eldius/document-feeder/internal/model"
	"github.com/eldius/initial-config-go/http/client"
	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

var (
	_ Parser = &feedParser{}
)

type Parser interface {
	Parse(ctx context.Context, feedURL string) (*model.Feed, error)
}

type feedParser struct {
	p *gofeed.Parser
	c *http.Client
}

func NewParser() Parser {
	p := gofeed.NewParser()
	c := client.NewHTTPClient()
	return &feedParser{p: p, c: c}
}

func (p *feedParser) Parse(ctx context.Context, feedURL string) (*model.Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "rss-parser/1.0.0 (+https://github.com/eldius/document-feeder)")

	res, err := p.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	feed, err := p.p.Parse(res.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing feed: %w", err)
	}

	return &model.Feed{
		Title:       feed.Title,
		Description: feed.Description,
		Link:        feed.Link,
		FeedLink:    feedURL,
		Links:       feed.Links,
		Language:    feed.Language,
		Image:       fromImage(feed.Image),
		Extensions:  fromExtensions(feed.Extensions),
		Items:       fromItems(feed.Items),
		FeedType:    feed.FeedType,
		FeedVersion: feed.FeedVersion,
	}, nil
}

func fromItems(items []*gofeed.Item) []model.Article {
	res := make([]model.Article, len(items))
	for i, item := range items {
		content := item.Content
		if content == "" {
			content = item.Description
		}
		res[i] = model.Article{
			Title:           item.Title,
			Description:     item.Description,
			Content:         content,
			Link:            item.Link,
			Links:           item.Links,
			Published:       item.Published,
			PublishedParsed: item.PublishedParsed,
			Authors:         fromAuthors(item.Authors),
			Guid:            item.GUID,
			DcExt:           model.DcExt{},
			Extensions:      model.ArticleExtension{},
			Image:           model.ArticleImage{},
			Categories:      item.Categories,
		}
	}
	return res
}

func fromAuthors(authors []*gofeed.Person) []model.Author {
	res := make([]model.Author, len(authors))
	for i, author := range authors {
		res[i] = model.Author{
			Name: author.Name,
		}
	}
	return res
}

func fromImage(img *gofeed.Image) model.Image {
	if img == nil {
		return model.Image{}
	}
	return model.Image{
		Url:   img.URL,
		Title: img.Title,
	}
}

func fromExtensions(ex ext.Extensions) model.Extensions {
	res := model.Extensions{}

	for k, v := range ex {
		res[k] = fromExtensionMap(v)
	}

	return res
}

func fromExtension(ex ext.Extension) model.Extension {
	return model.Extension{
		Name:     ex.Name,
		Value:    ex.Value,
		Attrs:    ex.Attrs,
		Children: fromExtensionMap(ex.Children),
	}
}

func fromExtensionMap(ex map[string][]ext.Extension) model.ExtensionMap {
	res := model.ExtensionMap{}
	for k, v := range ex {
		exs := make([]model.Extension, len(v))
		for i, e := range v {
			exs[i] = fromExtension(e)
		}
		res[k] = exs
	}

	return res
}

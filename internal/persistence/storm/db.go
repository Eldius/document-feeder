package storm

import (
	"context"
	"fmt"
	"os"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/eldius/document-feeder/internal/model"
	"github.com/eldius/initial-config-go/logs"
)

type Repository struct {
	db *storm.DB
}

func NewRepository() *Repository {
	_ = os.MkdirAll("data", 0700)
	db, _ := storm.Open("data/feeds.db")
	return &Repository{db: db}
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) Persist(_ context.Context, f *model.Feed) error {
	return r.db.Save(f)
}

func (r *Repository) All(_ context.Context) ([]*model.Feed, error) {
	var feeds []*model.Feed
	err := r.db.All(&feeds)
	return feeds, err
}

func (r *Repository) ArticleByLink(_ context.Context, feedTitle, articleLink string) (*model.Article, error) {
	log := logs.NewLogger(context.Background(), logs.KeyValueData{
		"feed_title":   feedTitle,
		"article_link": articleLink,
	})

	log.Debug("finding article by link")
	var feed model.Feed
	if err := r.db.Select(q.Eq("Title", feedTitle)).First(&feed); err != nil {
		return nil, fmt.Errorf("finding article by link: %w", err)
	}

	log = log.WithExtraData("feed_link", feed.FeedLink)
	for _, a := range feed.Items {
		log.WithExtraData("item_link", a.Link).Debug("checking article link")
		if a.Link == articleLink {
			return &a, nil
		}
	}
	return nil, nil
}

func (r *Repository) SaveGeneratedCache(_ context.Context, answer *model.AnswerCache) error {
	return r.db.Save(answer)
}

func (r *Repository) FindGeneratedCache(_ context.Context, id string) (*model.AnswerCache, error) {
	var answer model.AnswerCache
	if err := r.db.Select(q.Eq("ID", id)).First(&answer); err != nil {
		return nil, fmt.Errorf("finding generated cache: %w", err)
	}
	return &answer, nil
}

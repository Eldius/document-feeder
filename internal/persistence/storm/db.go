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

type Repository interface {
	Close() error
	Persist(context.Context, *model.Feed) error
	All(context.Context) ([]*model.Feed, error)
	ArticleByLink(context.Context, string, string) (*model.Article, error)
	SaveGeneratedCache(context.Context, *model.AnswerCache) error
	FindGeneratedCache(context.Context, string) (*model.AnswerCache, error)
}

type repository struct {
	db *storm.DB
}

func NewRepository() Repository {
	_ = os.MkdirAll("data", 0700)
	db, _ := storm.Open("data/feeds.db")
	return &repository{db: db}
}

func (r *repository) Close() error {
	return r.db.Close()
}

func (r *repository) Persist(_ context.Context, f *model.Feed) error {
	return r.db.Save(f)
}

func (r *repository) All(_ context.Context) ([]*model.Feed, error) {
	var feeds []*model.Feed
	err := r.db.All(&feeds)
	return feeds, err
}

func (r *repository) ArticleByLink(_ context.Context, feedTitle, articleLink string) (*model.Article, error) {
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

func (r *repository) SaveGeneratedCache(_ context.Context, answer *model.AnswerCache) error {
	return r.db.Save(answer)
}

func (r *repository) FindGeneratedCache(_ context.Context, id string) (*model.AnswerCache, error) {
	var answer model.AnswerCache
	if err := r.db.Select(q.Eq("ID", id)).First(&answer); err != nil {
		return nil, fmt.Errorf("finding generated cache: %w", err)
	}
	return &answer, nil
}

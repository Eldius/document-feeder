package storm

import (
	"context"
	"fmt"
	"go.etcd.io/bbolt"
	"os"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/eldius/document-feeder/internal/model"
	"github.com/eldius/initial-config-go/logs"
)

const (
	dbFileMode = 0766
)

type Repository interface {
	Close() error
	Persist(context.Context, *model.Feed) error
	All(context.Context) ([]*model.Feed, error)
	ArticleByLink(context.Context, string, string) (*model.Article, error)
	SaveGeneratedCache(context.Context, *model.AnswerCache) error
	FindGeneratedCache(context.Context, string) (*model.AnswerCache, error)
	Delete(context.Context, *model.Feed) error
}

type repository struct {
	db *storm.DB
}

func NewRepositoryFromDB(db *storm.DB) (Repository, error) {
	return &repository{db: db}, nil
}

func NewRepository() (Repository, error) {

	_ = os.MkdirAll("data", dbFileMode)

	db, err := storm.Open(
		"data/feeds.db",
		storm.BoltOptions(dbFileMode, &bbolt.Options{
			Timeout: 3 * time.Second,
		}),
		storm.Batch(),
	)
	if err != nil {
		err := fmt.Errorf("opening db: %w", err)
		fmt.Println("Failed opening db:", err)
		return nil, err
	}
	return NewRepositoryFromDB(db)
}

func (r *repository) Close() error {
	return r.db.Close()
}

func (r *repository) Persist(_ context.Context, f *model.Feed) error {
	return r.db.Save(f)
}

func (r *repository) All(_ context.Context) ([]*model.Feed, error) {
	if r.db == nil {
		return nil, fmt.Errorf("db is nil")
	}
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

func (r *repository) Delete(_ context.Context, feed *model.Feed) error {
	return r.db.DeleteStruct(feed)
}

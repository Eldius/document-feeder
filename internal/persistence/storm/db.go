package storm

import (
	"context"
	"github.com/asdine/storm/v3"
	"github.com/eldius/document-feed-embedder/internal/model"
	"os"
)

type Repository struct {
	db *storm.DB
}

func NewRepository() *Repository {
	_ = os.MkdirAll("data", 0755)
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

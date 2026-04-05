package chromem

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/asdine/storm/v3"
	"github.com/eldius/document-feeder/internal/client/ollama"
	"github.com/eldius/document-feeder/internal/config"
	"github.com/eldius/document-feeder/internal/model"
	"github.com/eldius/initial-config-go/logs"
	"github.com/philippgille/chromem-go"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

const (
	articleCollectionName = "articles"
	answerCollectionName  = "answers"
)

type DocumentVectorizer interface {
	Search(ctx context.Context, term string, maxResults int) ([]chromem.Result, error)
	SearchWithSimilarityFilter(ctx context.Context, term string, maxResults int, similarity float32) ([]chromem.Result, error)
	Save(ctx context.Context, feed *model.Feed) error
	SaveGenerationCache(ctx context.Context, cache *model.AnswerCache) (string, error)
	FindCacheID(ctx context.Context, question string) (string, error)
	DeleteByFeedLink(ctx context.Context, feedLink string) error
	Delete(ctx context.Context, ids ...string) error
}

type vectorizer struct {
	db                *chromem.DB
	docsCollection    *chromem.Collection
	answersCollection *chromem.Collection

	embeddingModel string
	chunkSize      int
	chunkOverlap   int
	textSplitter   textsplitter.TextSplitter
}

func NewDocumentVectorizer(
	db *chromem.DB,
	textSplitter textsplitter.TextSplitter,
	ollamaClient ollama.Client,
	embeddingModel string,
	chunkSize, chunkOverlap int,
) (DocumentVectorizer, error) {
	docsCollection, err := db.GetOrCreateCollection(articleCollectionName, map[string]string{}, NewEmbeddingFuncOllama(embeddingModel, ollamaClient))
	if err != nil {
		return nil, err
	}

	answersCollection, err := db.GetOrCreateCollection(answerCollectionName, map[string]string{}, NewEmbeddingFuncOllama(embeddingModel, ollamaClient))
	if err != nil {
		return nil, err
	}

	de := &vectorizer{
		db:                db,
		embeddingModel:    embeddingModel,
		chunkSize:         chunkSize,
		chunkOverlap:      chunkOverlap,
		docsCollection:    docsCollection,
		answersCollection: answersCollection,
		textSplitter:      textSplitter,
	}

	return de, nil
}

func NewDefaultDocumentVectorizer() (DocumentVectorizer, error) {
	embeddingModel := config.GetOllamaEmbeddingModel()

	dbPath := fmt.Sprintf("data/doc_%s.db", embeddingModel)
	dbPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("getting absolute path: %w", err)
	}

	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := chromem.NewPersistentDB(dbPath, true)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	splitter := textsplitter.NewTokenSplitter(
		textsplitter.WithChunkSize(config.GetOllamaEmbeddingChunkSize()),
		textsplitter.WithModelName(embeddingModel),
		textsplitter.WithChunkOverlap(config.GetOllamaEmbeddingChunkOverlap()),
	)

	ollamaClient, err := ollama.NewOllamaClientFromConfigs()
	if err != nil {
		return nil, fmt.Errorf("creating ollama client: %w", err)
	}
	return NewDocumentVectorizer(
		db,
		splitter,
		ollamaClient,
		embeddingModel,
		config.GetOllamaEmbeddingChunkSize(),
		config.GetOllamaEmbeddingChunkOverlap(),
	)
}

func (d *vectorizer) Search(ctx context.Context, term string, maxResults int) ([]chromem.Result, error) {
	log := logs.NewLogger(ctx, logs.KeyValueData{
		"term":        term,
		"max_results": maxResults,
	})
	log.Debug("searching documents")
	results, err := d.docsCollection.Query(ctx, term, maxResults, nil, nil)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			log.WithError(err).Warn("no results found")
			return nil, nil
		}
		return nil, err
	}
	log.WithExtraData("results", results).
		WithExtraData("results_count", len(results)).
		Debug("doc results")
	return results, err
}

func (d *vectorizer) SearchDocs(ctx context.Context, term string, maxResults int) ([]chromem.Result, error) {
	return d.docsCollection.Query(ctx, term, maxResults, nil, nil)
}

func (d *vectorizer) Save(ctx context.Context, feed *model.Feed) error {
	if feed == nil {
		return nil
	}
	log := logs.NewLogger(ctx, logs.KeyValueData{
		"feed_title": feed.Title,
	})
	log.Debug("saving feed")
	var docs []chromem.Document
	for _, article := range feed.Items {

		if article.Content == "" {
			continue
		}

		splitDocs, err := d.htmlParse(ctx, article)
		if err != nil {
			logs.NewLogger(ctx, logs.KeyValueData{
				"error": err,
			}).Warnf("error parsing article")
			continue
		}

		metadata := map[string]string{
			"title":     article.Title,
			"link":      article.Link,
			"feed":      feed.Title,
			"feed_link": feed.FeedLink,
			"date":      article.PublishedParsed.Format("2006-01-02"),
			"tags":      strings.Join(article.Categories, ","),
		}
		for i, doc := range splitDocs {
			md := maps.Clone(metadata)
			chunkID := fmt.Sprintf("%.0000d-", i) + article.Link
			md["chunk_id"] = chunkID
			md["score"] = fmt.Sprintf("%.3f", doc.Score)
			docs = append(docs, chromem.Document{
				ID:       chunkID + article.Link,
				Metadata: md,
				Content:  doc.PageContent,
			})
		}
	}
	if len(docs) == 0 {
		return nil
	}

	if err := d.docsCollection.AddDocuments(ctx, docs, runtime.NumCPU()); err != nil {
		err := fmt.Errorf("adding documents: %w", err)
		log.WithExtraData("error", err).Warn("error adding document")
		return err
	}
	return nil
}

func (d *vectorizer) SaveGenerationCache(ctx context.Context, cache *model.AnswerCache) (string, error) {
	id := cacheID(cache)
	docsToAdd := []chromem.Document{{
		ID:       id,
		Metadata: map[string]string{"question": cache.Question},
		Content:  cache.Answer,
	}}
	if err := d.answersCollection.AddDocuments(ctx, docsToAdd, runtime.NumCPU()); err != nil {
		return "", fmt.Errorf("adding document: %w", err)
	}
	return id, nil
}

func (d *vectorizer) FindCacheID(ctx context.Context, question string) (string, error) {
	queryResults, err := d.answersCollection.Query(ctx, question, 1, nil, nil)
	if err != nil {
		return "", fmt.Errorf("querying documents: %w", err)
	}
	if len(queryResults) == 0 {
		return "", nil
	}
	return queryResults[0].ID, nil
}

func (d *vectorizer) SearchWithSimilarityFilter(ctx context.Context, term string, maxResults int, similarity float32) ([]chromem.Result, error) {
	docs, err := d.Search(ctx, term, maxResults)
	if err != nil {
		return nil, err
	}
	var filteredDocs []chromem.Result
	for _, doc := range docs {
		if doc.Similarity >= similarity {
			filteredDocs = append(filteredDocs, doc)
		}
	}
	return filteredDocs, nil
}

func (d *vectorizer) htmlParse(ctx context.Context, article model.Article) ([]schema.Document, error) {
	return documentloaders.NewHTML(strings.NewReader(article.Content)).
		LoadAndSplit(
			ctx,
			d.textSplitter,
		)
}

func cacheID(cache *model.AnswerCache) string {
	h := sha256.New()
	h.Write([]byte(cache.Question))
	h.Write([]byte(cache.Answer))
	return hex.EncodeToString(h.Sum(nil))
}

func (d *vectorizer) DeleteByFeedLink(ctx context.Context, feedLink string) error {
	docs, err := d.docsCollection.QueryWithOptions(ctx, chromem.QueryOptions{
		QueryText:      "a",
		QueryEmbedding: make([]float32, 0),
		NResults:       1000,
		Where: map[string]string{
			"feed_link": feedLink,
		},
	})
	if err != nil {
		return fmt.Errorf("querying documents for deletion: %w", err)
	}
	var ids []string
	for _, doc := range docs {
		fmt.Println("-", doc.ID)
		ids = append(ids, doc.ID)
	}
	if err := d.Delete(ctx, ids...); err != nil {
		return fmt.Errorf("deleting documents: %w", err)
	}
	return err
}

func (d *vectorizer) Delete(ctx context.Context, ids ...string) error {
	if len(ids) == 0 {
		return nil
	}
	return d.docsCollection.Delete(ctx, nil, nil, ids...)
}

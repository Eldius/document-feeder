package chromem

import (
	"context"
	"fmt"
	"github.com/eldius/document-feed-embedder/internal/client"
	"github.com/eldius/document-feed-embedder/internal/config"
	"github.com/eldius/document-feed-embedder/internal/model"
	"github.com/eldius/initial-config-go/logs"
	"github.com/philippgille/chromem-go"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"maps"
	"strings"
)

const articleCollectionName = "articles"

type DocumentVectorizer struct {
	db           *chromem.DB
	coll         *chromem.Collection
	ollamaClient *client.OllamaClient

	embeddingModel string
	chunkSize      int
	chunkOverlap   int
}

type ollamaResponse struct {
	Model           string      `json:"model"`
	Embeddings      [][]float32 `json:"embeddings"`
	TotalDuration   int64       `json:"total_duration"`
	LoadDuration    int         `json:"load_duration"`
	PromptEvalCount int         `json:"prompt_eval_count"`
}

func NewDocumentVectorizer(db *chromem.DB, oc *client.OllamaClient, embeddingFunc chromem.EmbeddingFunc, embeddingModel string, chunkSize, chunkOverlap int) (*DocumentVectorizer, error) {
	coll, err := db.GetOrCreateCollection(articleCollectionName, map[string]string{}, embeddingFunc)
	if err != nil {
		return nil, err
	}

	de := &DocumentVectorizer{
		db:             db,
		ollamaClient:   oc,
		embeddingModel: embeddingModel,
		chunkSize:      chunkSize,
		chunkOverlap:   chunkOverlap,
		coll:           coll,
	}

	return de, nil
}

func NewDefaultDocumentVectorizer() (*DocumentVectorizer, error) {
	db, err := chromem.NewPersistentDB("data/doc.db", true)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	o := client.NewOllamaClient()
	return NewDocumentVectorizer(db, o, o.EmbeddingFunc, config.GetOllamaEmbeddingModel(), config.GetOllamaEmbeddingBatchSize(), config.GetOllamaEmbeddingChunkOverlap())
}

func (d *DocumentVectorizer) Search(ctx context.Context, term string, maxResults int) ([]chromem.Result, error) {
	fmt.Println("searching for term:", term)
	embTerm, err := d.ollamaClient.EmbeddingFunc(ctx, term)
	if err != nil {
		return nil, fmt.Errorf("embedding term: %w", err)
	}
	//queryText := "What are the animals doing?"
	queryResults, err := d.coll.QueryEmbedding(ctx, embTerm, maxResults, nil, nil)

	return queryResults, nil
}

func (d *DocumentVectorizer) SearchDocs(ctx context.Context, term string, maxResults int) ([]chromem.Result, error) {
	fmt.Println("searching for term:", term)
	embTerm, err := d.ollamaClient.EmbeddingFunc(ctx, term)
	if err != nil {
		return nil, fmt.Errorf("embedding term: %w", err)
	}
	//queryText := "What are the animals doing?"
	queryResults, err := d.coll.QueryEmbedding(ctx, embTerm, maxResults, nil, nil)

	return queryResults, nil
}

func (d *DocumentVectorizer) Save(ctx context.Context, feed *model.Feed) error {
	if feed == nil {
		fmt.Println("- no feed to save")
		return nil
	}
	fmt.Println("- processing feed:", feed.Title)

	var docs []chromem.Document
	for _, article := range feed.Items {

		fmt.Println("  - processing article:", article.Title, "-", len(article.Content), "chars")
		if article.Content == "" {
			fmt.Println("  - skipping article without content:", article.Title)
			continue
		}

		spltDocs, err := d.htmlParse(ctx, article)
		if err != nil {
			logs.NewLogger(ctx, logs.KeyValueData{
				"error": err,
			}).Warnf("error parsing article")
			continue
		}

		var embDocs []chromem.Document
		docCount := len(spltDocs)
		metadata := map[string]string{
			"title":     article.Title,
			"link":      article.Link,
			"feed":      feed.Title,
			"feed_link": feed.FeedLink,
			"date":      article.PublishedParsed.Format("2006-01-02"),
			"tags":      strings.Join(article.Categories, ","),
		}
		for i, doc := range spltDocs {
			fmt.Printf("  - processing chunk %d/%d => (%s) %d chars long\n", i+1, docCount, article.Title, len(doc.PageContent))
			embContent, err := d.ollamaClient.EmbeddingFunc(ctx, doc.PageContent)
			if err != nil {
				logs.NewLogger(ctx, logs.KeyValueData{
					"error": err,
				}).Warnf("error embedding document")
				continue
			}
			md := maps.Clone(metadata)
			md["chunk"] = fmt.Sprintf("%d/%d", i+1, docCount)
			md["chunk_size"] = fmt.Sprintf("%d", len(doc.PageContent))
			embDoc, err := chromem.NewDocument(ctx, article.Link, md, embContent, doc.PageContent, d.ollamaClient.EmbeddingFunc)
			if err != nil {
				logs.NewLogger(ctx, logs.KeyValueData{
					"error": err,
				}).Warnf("error creating document")
				continue
			}
			embDocs = append(embDocs, embDoc)
		}

		docs = append(docs, embDocs...)
	}
	if len(docs) == 0 {
		return nil
	}

	if err := d.coll.AddDocuments(ctx, docs, 1); err != nil {
		logs.NewLogger(ctx, logs.KeyValueData{
			"error": err,
		}).Warnf("error adding document")
		return err
	}
	return nil
}

func (d *DocumentVectorizer) htmlParse(ctx context.Context, article model.Article) ([]schema.Document, error) {
	fmt.Println()
	fmt.Println("---")
	fmt.Println("  - [htmlParse] processing article:", article.Title)
	html := documentloaders.NewHTML(strings.NewReader(article.Content))
	return html.LoadAndSplit(
		ctx,
		textsplitter.NewTokenSplitter(
			textsplitter.WithChunkSize(d.chunkSize),
			textsplitter.WithModelName(d.embeddingModel),
			textsplitter.WithChunkOverlap(d.chunkOverlap),
		),
	)
}

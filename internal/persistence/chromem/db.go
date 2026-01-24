package chromem

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/eldius/document-feed-embedder/internal/config"
	"github.com/eldius/document-feed-embedder/internal/model"
	"github.com/eldius/initial-config-go/httpclient"
	"github.com/eldius/initial-config-go/logs"
	"github.com/philippgille/chromem-go"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"maps"
	"net/http"
	"strings"
)

type DocumentEmbedder struct {
	db *chromem.DB
	c  *http.Client

	embeddingModel string
	ollamaEndpoint string
	chunkSize      int
	chunkOverlap   int
}

type ollamaRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaResponse struct {
	Model           string      `json:"model"`
	Embeddings      [][]float32 `json:"embeddings"`
	TotalDuration   int64       `json:"total_duration"`
	LoadDuration    int         `json:"load_duration"`
	PromptEvalCount int         `json:"prompt_eval_count"`
}

func NewDocumentEmbedder() (*DocumentEmbedder, error) {
	db, err := chromem.NewPersistentDB("data/doc.db", true)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}
	c := httpclient.NewHTTPClient()
	return &DocumentEmbedder{
		db:             db,
		c:              c,
		embeddingModel: config.GetOllamaEmbeddingModel(),
		ollamaEndpoint: config.GetOllamaEndpoint(),
		chunkSize:      config.GetOllamaEmbeddingBatchSize(),
		chunkOverlap:   config.GetOllamaEmbeddingChunkOverlap(),
	}, nil
}

func (d *DocumentEmbedder) EmbeddingFunction(ctx context.Context, text string) ([]float32, error) {

	reqPayload := ollamaRequest{
		Model: d.embeddingModel,
		Input: []string{text},
	}

	b, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, d.ollamaEndpoint+"/api/embed", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	res, err := d.c.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer func() { _ = res.Body.Close() }()

	var embeddings ollamaResponse
	if err := json.NewDecoder(res.Body).Decode(&embeddings); err != nil {
		return nil, err
	}

	return embeddings.Embeddings[0], nil
}

func (d *DocumentEmbedder) Search(ctx context.Context, term string) ([]*model.Article, error) {
	fmt.Println("searching for term:", term)
	embTerm, err := d.EmbeddingFunction(ctx, term)
	if err != nil {
		return nil, fmt.Errorf("embedding term: %w", err)
	}
	fmt.Println("term embedding:", embTerm)
	//d.db.
	return nil, nil
}

func (d *DocumentEmbedder) Save(ctx context.Context, feed *model.Feed) error {
	if feed == nil {
		fmt.Println("- no feed to save")
		return nil
	}
	fmt.Println("- processing feed:", feed.Title)

	coll, err := d.db.GetOrCreateCollection("articles", map[string]string{}, d.EmbeddingFunction)
	if err != nil {
		return err
	}

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
			"title": article.Title,
			"link":  article.Link,
			"feed":  feed.Title,
			"date":  article.PublishedParsed.Format("2006-01-02"),
			"tags":  strings.Join(article.Categories, ","),
		}
		for i, doc := range spltDocs {
			fmt.Printf("  - processing chunk %d/%d => (%s) %d chars long\n", i+1, docCount, article.Title, len(doc.PageContent))
			embContent, err := d.EmbeddingFunction(ctx, doc.PageContent)
			if err != nil {
				logs.NewLogger(ctx, logs.KeyValueData{
					"error": err,
				}).Warnf("error embedding document")
				continue
			}
			md := maps.Clone(metadata)
			md["chunk"] = fmt.Sprintf("%d/%d", i+1, docCount)
			md["chunk_size"] = fmt.Sprintf("%d", len(doc.PageContent))
			embDoc, err := chromem.NewDocument(ctx, article.Link, md, embContent, doc.PageContent, d.EmbeddingFunction)
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

	if err := coll.AddDocuments(ctx, docs, 1); err != nil {
		logs.NewLogger(ctx, logs.KeyValueData{
			"error": err,
		}).Warnf("error adding document")
		return err
	}
	return nil
}

func (d *DocumentEmbedder) htmlParse(ctx context.Context, article model.Article) ([]schema.Document, error) {
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

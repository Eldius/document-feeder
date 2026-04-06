package adapter

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/eldius/initial-config-go/logs"
	chromem2 "github.com/philippgille/chromem-go"

	"github.com/eldius/document-feeder/internal/client/ollama"
	"github.com/eldius/document-feeder/internal/config"
	"github.com/eldius/document-feeder/internal/feed"
	"github.com/eldius/document-feeder/internal/model"
	"github.com/eldius/document-feeder/internal/persistence/chromem"
	"github.com/eldius/document-feeder/internal/persistence/storm"
)

var (
	//go:embed templates/**
	templates embed.FS
)

type FeedAdapter struct {
	r          storm.Repository
	docs       chromem.DocumentVectorizer
	p          feed.Parser
	tmpl       *template.Template
	ollama     ollama.Client
	cache      cacheConfig
	contextCfg contextConfig
	models     *feedAdapterModels
}

type feedAdapterModels struct {
	articleAnalysisModel string
	generationModel      string
	embeddingModel       string
}

type cacheConfig struct {
	Enabled             bool
	SimilarityThreshold float32
}
type contextConfig struct {
	SimilarityThreshold float32
	Enabled             bool
	MaxDocuments        int
}

func NewFeedAdapter(
	r storm.Repository,
	p feed.Parser,
	docs chromem.DocumentVectorizer,
	ollama ollama.Client,
	tmpl *template.Template,
	cacheSimilarityThreshold float32,
	cacheEnabled bool,
	contextSimilarityThreshold float32,
	contextEnabled bool,
	contextMaxDocuments int,
	models *feedAdapterModels,
) *FeedAdapter {
	return &FeedAdapter{
		r:      r,
		p:      p,
		docs:   docs,
		tmpl:   tmpl,
		ollama: ollama,
		models: models,
		cache: cacheConfig{
			Enabled:             cacheEnabled,
			SimilarityThreshold: cacheSimilarityThreshold,
		},
		contextCfg: contextConfig{
			SimilarityThreshold: contextSimilarityThreshold,
			Enabled:             contextEnabled,
			MaxDocuments:        contextMaxDocuments,
		},
	}
}

func NewFeedAdapterFromConfigs() (*FeedAdapter, error) {
	r, err := storm.NewRepository()
	if err != nil {
		return nil, fmt.Errorf("creating repository: %w", err)
	}
	p := feed.NewParser()
	d, err := chromem.NewDefaultDocumentVectorizer()
	if err != nil {
		return nil, fmt.Errorf("creating document embedder: %w", err)
	}
	tmpl, err := template.ParseFS(templates, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}
	o, err := ollama.NewOllamaClientFromConfigs()
	if err != nil {
		return nil, fmt.Errorf("creating ollama client: %w", err)
	}

	cacheEnabled := config.GetOllamaGenerationCacheEnabled()
	cacheSimilarityThreshold := config.GetOllamaGenerationCacheSimilarityThreshold()

	contextEnabled := config.GetOllamaGenerationContextEnabled()
	contextMaxDocuments := config.GetOllamaGenerationContextMaxDocuments()
	contextSimilarityThreshold := config.GetOllamaGenerationContextSimilarityThreshold()

	models := &feedAdapterModels{
		generationModel:      config.GetOllamaGenerationModel(),
		embeddingModel:       config.GetOllamaEmbeddingModel(),
		articleAnalysisModel: config.GetOllamaArticleAnalysisModel(),
	}
	return NewFeedAdapter(r, p, d, o, tmpl, cacheSimilarityThreshold, cacheEnabled, contextSimilarityThreshold, contextEnabled, contextMaxDocuments, models), nil
}

func (a *FeedAdapter) Parse(ctx context.Context, feedURL string) (*model.Feed, error) {
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
		if err := a.RefreshFeed(ctx, f); err != nil {
			return fmt.Errorf("refreshing feed: %w", err)
		}
	}
	return nil
}

func (a *FeedAdapter) RefreshFeed(ctx context.Context, f *model.Feed) error {
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
	return nil
}

func (a *FeedAdapter) All(ctx context.Context) ([]*model.Feed, error) {
	return a.r.All(ctx)
}

func (a *FeedAdapter) SearchWithSimilarityThreshold(ctx context.Context, term string, maxResults int, similarity float32) ([]*model.SearchResult, error) {
	docs, err := a.docs.SearchWithSimilarityFilter(ctx, term, maxResults, similarity)
	if err != nil {
		return nil, fmt.Errorf("searching documents: %w", err)
	}

	var res model.SearchResultList
	for _, d := range docs {
		sr, err := a.documentReverseSearch(ctx, d)
		if err != nil {
			return nil, fmt.Errorf("searching documents: %w", err)
		}
		res.Results = append(res.Results, sr)
	}

	return res.Sorted(), nil
}

func (a *FeedAdapter) Search(ctx context.Context, term string, maxResults int) ([]*model.SearchResult, error) {
	docs, err := a.docs.Search(ctx, term, maxResults)
	if err != nil {
		return nil, fmt.Errorf("searching documents: %w", err)
	}

	var res model.SearchResultList
	for _, d := range docs {
		if d.Similarity < a.contextCfg.SimilarityThreshold {
			continue
		}
		sr, err := a.documentReverseSearch(ctx, d)
		if err != nil {
			return nil, fmt.Errorf("searching documents: %w", err)
		}
		res.Results = append(res.Results, sr)
	}

	return res.Sorted(), nil
}

func (a *FeedAdapter) documentReverseSearch(ctx context.Context, d chromem2.Result) (*model.SearchResult, error) {
	doc, err := a.r.ArticleByLink(ctx, d.Metadata["feed"], d.Metadata["link"])
	if err != nil {
		return nil, fmt.Errorf("getting article by link: %w", err)
	}
	if doc == nil {
		return &model.SearchResult{
			FeedTitle:        d.Metadata["feed"],
			Similarity:       d.Similarity,
			SanitizedContent: d.Content,
			Embeddings:       d.Embedding,
		}, nil
	}
	return &model.SearchResult{
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
	}, nil
}

type promptTemplateData struct {
	Question string
	Results  []*model.SearchResult
}

func (a *FeedAdapter) AskAQuestion(ctx context.Context, question string) (string, error) {
	if a.cache.Enabled {
		cacheID, err := a.docs.FindCacheID(ctx, question)
		if err == nil && cacheID != "" {
			cache, err := a.r.FindGeneratedCache(ctx, cacheID)
			if err != nil {
				return "", fmt.Errorf("finding generated cache: %w", err)
			}

			return cache.Answer, nil
		}
	}
	var docs []*model.SearchResult
	if a.contextCfg.Enabled {
		var err error
		docs, err = a.SearchWithSimilarityThreshold(ctx, question, a.contextCfg.MaxDocuments, a.contextCfg.SimilarityThreshold)
		if err != nil {
			return "", fmt.Errorf("searching documents: %w", err)
		}
	}

	data := promptTemplateData{Question: question, Results: docs}

	var b strings.Builder
	if err := a.tmpl.ExecuteTemplate(&b, "prompt.tmpl", data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	response, err := a.ollama.GenerateFunc(
		ctx,
		b.String(),
		ollama.WithNumPredict(1024),
	)
	if err != nil {
		return "", fmt.Errorf("generating response: %w", err)
	}

	if a.cache.Enabled {
		cache := model.AnswerCache{
			Question: question,
			Answer:   response.Response,
		}
		cacheID, err := a.docs.SaveGenerationCache(ctx, &cache)
		if err != nil {
			return "", fmt.Errorf("saving answer cache: %w", err)
		}
		cache.ID = cacheID
		if err := a.r.SaveGeneratedCache(ctx, &cache); err != nil {
			return "", fmt.Errorf("saving answer cache: %w", err)
		}
	}

	return response.Response, err
}

func (a *FeedAdapter) AskAQuestionStream(ctx context.Context, question string, ch chan string) error {
	if a.cache.Enabled {
		cacheID, err := a.docs.FindCacheID(ctx, question)
		if err == nil && cacheID != "" {
			cache, err := a.r.FindGeneratedCache(ctx, cacheID)
			if err == nil {
				ch <- cache.Answer
				return nil
			}
			logs.NewLogger(ctx).WithError(err).Warn("error finding generated cache")
		}
	}
	docs, err := a.Search(ctx, question, 2)
	if err != nil {
		return fmt.Errorf("searching documents: %w", err)
	}

	data := promptTemplateData{Question: question, Results: docs}

	var b strings.Builder
	if err := a.tmpl.ExecuteTemplate(&b, "prompt.tmpl", data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	if err := a.ollama.GenerateCallStream(
		ctx,
		ch,
		ollama.GenerateRequest{
			Prompt:    b.String(),
			KeepAlive: 0,
		},
	); err != nil {
		return fmt.Errorf("generating response: %w", err)
	}

	return nil
}

func (a *FeedAdapter) SanitizeArticlesDB(ctx context.Context) error {
	feeds, err := a.r.All(ctx)
	if err != nil {
		return fmt.Errorf("getting all feeds: %w", err)
	}
	for _, f := range feeds {
		log := logs.NewLogger(ctx).WithExtraData("feed", f.FeedLink)
		if strings.TrimSpace(f.FeedLink) == "" {
			if err := a.deleteFeed(ctx, f); err != nil {
				log.WithError(err).WithExtraData("feed", f.FeedLink).Warn("deleting feed with empty feed link")
				return err
			}
			log.Debug("deleted feed with empty title")
			continue
		}

		if strings.TrimSpace(f.Title) == "" {
			if err := a.deleteFeed(ctx, f); err != nil {
				log.WithError(err).WithExtraData("feed", f.FeedLink).Warn("deleting feed with empty title")
				return err
			}
			if err := a.deleteFeedEmbeddings(ctx, f); err != nil {
				log.WithError(err).WithExtraData("feed", f.FeedLink).Warn("deleting embeddings for feed with empty title")
				return err
			}
			log.Warn("deleted feed with empty title")
		}
		if err := a.r.Persist(ctx, f); err != nil {
			return fmt.Errorf("persisting feed: %w", err)
		}
	}
	return nil
}

func (a *FeedAdapter) deleteFeedEmbeddings(ctx context.Context, f *model.Feed) error {
	if err := a.docs.DeleteByFeedLink(ctx, f.FeedLink); err != nil {
		err = fmt.Errorf("deleting embeddings for feed: %w", err)
		return err
	}
	return nil
}

func (a *FeedAdapter) deleteFeed(ctx context.Context, f *model.Feed) error {
	if err := a.r.Delete(ctx, f); err != nil {
		err = fmt.Errorf("deleting feeds: %w", err)
		return err
	}
	return nil
}

func (a *FeedAdapter) AnalyzeArticle(ctx context.Context, article model.Article, skipCache bool) (*model.AnalysisResult, error) {
	if !skipCache {
		analysis, err := a.r.FindArticleAnalysis(ctx, article.Link)
		if err == nil && analysis != nil {
			return &analysis.Analysis, nil
		}
	}

	var b strings.Builder
	if err := a.tmpl.ExecuteTemplate(&b, "analyze.tmpl", article); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	response, err := a.ollama.GenerateCall(ctx, ollama.GenerateRequest{
		Prompt: b.String(),
		Format: "json",
		Model:  a.models.articleAnalysisModel,
		Options: ollama.OptionsRequest{
			Temperature: 0.1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("generating analysis: %w", err)
	}

	var result model.AnalysisResult
	if err := json.Unmarshal([]byte(response.Response), &result); err != nil {
		return nil, fmt.Errorf("unmarshalling analysis: %w", err)
	}

	if err := a.r.SaveArticleAnalysis(ctx, &model.ArticleAnalysis{
		ArticleLink: article.Link,
		Analysis:    result,
	}); err != nil {
		logs.NewLogger(ctx).WithError(err).Warn("failed to save article analysis")
	}

	return &result, nil
}

func (a *FeedAdapter) AnalyzeFeed(ctx context.Context, feedLink string, skipCache bool) (map[string]*model.AnalysisResult, error) {
	feeds, err := a.r.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting all feeds: %w", err)
	}

	var f *model.Feed
	for _, feed_ := range feeds {
		if feed_.FeedLink == feedLink {
			f = feed_
			break
		}
	}

	if f == nil {
		return nil, fmt.Errorf("feed not found: %s", feedLink)
	}

	results := make(map[string]*model.AnalysisResult)
	for _, item := range f.Items {
		analysis, err := a.AnalyzeArticle(ctx, item, skipCache)
		if err != nil {
			logs.NewLogger(ctx).WithError(err).WithExtraData("article", item.Link).Warn("failed to analyze article")
			continue
		}
		results[item.Link] = analysis
	}

	return results, nil
}

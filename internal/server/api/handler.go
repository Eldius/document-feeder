package api

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/initial-config-go/http/server"
	"github.com/eldius/initial-config-go/logs"
)

type handler struct {
	feedAdapter *adapter.FeedAdapter
	notifier    adapter.Notifier
	staticFiles fs.FS
}

var (
	//go:embed static/*
	staticFiles embed.FS
)

func (h *handler) listFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.feedAdapter.All(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "plain/text")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "error: %s", err)
		return
	}

	var resp []FeedSummary
	for _, f := range feeds {
		resp = append(resp, FeedSummary{
			Title: f.Title,
			URL:   f.Link,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "error: %s", err)
		return
	}
}

func (h *handler) refreshFeeds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logs.NewLogger(ctx)
	log.Info("refreshing feeds")

	w.Header().Set("Content-Type", "application/json")

	flusher, ok := w.(http.Flusher)
	if !ok {
		err := h.feedAdapter.Refresh(ctx)
		if err != nil {
			log.WithError(err).Error("failed to get feeds")
			if err := h.notifier.Notify(ctx, fmt.Sprintf("Failed to get feeds: %s", err)); err != nil {
				log.WithError(err).Error("failed to notify")
			}
			http.Error(w, fmt.Sprintf("failed to get feeds: %s", err), http.StatusInternalServerError)
			return
		}
		return
	}

	encoder := json.NewEncoder(w)
	all, err := h.feedAdapter.All(ctx)
	if err != nil {
		log.WithError(err).Error("failed to get feeds")
		if err := h.notifier.Notify(ctx, fmt.Sprintf("Failed to get feeds: %s", err)); err != nil {
			log.WithError(err).Error("failed to notify")
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "error: %s", err)
		return
	}

	log = log.WithExtraData("feeds_count", len(all))

	log.Info("feeds retrieved")

	w.WriteHeader(http.StatusContinue)
	var res []FeedSummary
	for _, f := range all {
		if err := h.feedAdapter.RefreshFeed(ctx, f); err != nil {
			log.WithError(err).WithExtraData("feed_name", f.Title).Error("failed to refresh feed")
			_, _ = fmt.Fprintf(w, "failed to refresh feed %s: %s", f.Title, err)
			continue
		}
		summary := ToFeedSummary(f)
		_ = encoder.Encode(summary)
		res = append(res, *summary)
		flusher.Flush()
	}

	if err := h.notifier.Notify(ctx, "Feeds refreshed"); err != nil {
		log.WithError(err).Error("failed to notify")
	}

	log.Info("feeds refreshed")

	w.WriteHeader(http.StatusOK)
	if err := encoder.Encode(&res); err != nil {
		log.WithError(err).Error("failed to encode feed summary")
		_, _ = fmt.Fprint(w, "failed to encode feed summary: "+err.Error()+"\n"+
			"Please check the logs for more details.",
		)
	}
	flusher.Flush()
}

func (h *handler) addFeeds(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	ctx := r.Context()
	log := logs.NewLogger(ctx)
	log.Info("refreshing feeds")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")

	var req AddFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "error: %s", err)
		return
	}
	flusher := w.(http.Flusher)
	encoder := json.NewEncoder(w)

	if err := h.processAddFeeds(ctx, &req, encoder, flusher); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "error: %s", err)
	}
}

func (h *handler) searchOnFeeds(w http.ResponseWriter, r *http.Request) {

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusBadRequest)
		return
	}
	searchResult, err := h.feedAdapter.Search(r.Context(), req.Query, 10)
	if err != nil {
		http.Error(w, "error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var resp SearchResponse
	for _, f := range searchResult {
		resp.Results = append(resp.Results, SearchResult{
			FeedTitle: f.FeedTitle,
			Article: Article{
				Title:       f.Article.Title,
				Description: f.Article.Description,
				Content:     f.Article.Content,
				Link:        f.Article.Link,
			},
			Similarity: f.Similarity,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&resp)
}

func (h *handler) processAddFeeds(ctx context.Context, req *AddFeedRequest, encoder *json.Encoder, flusher http.Flusher) error {
	var wg sync.WaitGroup

	out := streamOutput{encoder: encoder, flusher: flusher}

	_ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()
	wg.Go(func() {
		ctx := _ctx
		log := logs.NewLogger(ctx)
		for {
			select {
			case <-ctx.Done():
				log.Info("context cancelled")
				return
			case <-time.After(10 * time.Second):
				log.Info("ping")
				out.partialOutput(FeedSummary{})
			}
		}
	})

	log := logs.NewLogger(ctx)
	var res []FeedSummary
	for _, f := range req.Feeds {
		feed, err := h.feedAdapter.Parse(ctx, f)
		if err != nil {
			log.WithError(err).WithExtraData("feed_name", f).Error("failed to refresh feed")
			summary := FeedSummary{
				Title: "",
				URL:   f,
				Error: err.Error(),
			}
			out.partialOutput(summary)
			res = append(res, summary)
			continue
		}

		summary := ToFeedSummary(feed)
		out.partialOutput(*summary)
		res = append(res, *summary)
	}
	cancelFunc()

	_ = encoder.Encode(&res)
	return nil
}

type streamOutput struct {
	encoder *json.Encoder
	flusher http.Flusher
	sync.Mutex
}

func (o *streamOutput) partialOutput(val any) {
	if o.encoder != nil {
		_ = o.encoder.Encode(val)
	}
	if o.flusher != nil {
		o.flusher.Flush()
	}
}

func newHandler(
	feedAdapter *adapter.FeedAdapter,
	notifier adapter.Notifier,
	fs fs.FS,
) *handler {
	return &handler{
		feedAdapter: feedAdapter,
		notifier:    notifier,
		staticFiles: fs,
	}
}

func StartServer(_ context.Context, port int) error {
	a, err := adapter.NewFeedAdapterFromConfigs()
	if err != nil {
		err := fmt.Errorf("creating adapter: %w", err)
		fmt.Printf("failed to create adapter: %s\n", err)
		return err
	}

	n := adapter.NewXmppNotifierFromConfigs()

	h := newHandler(a, n, staticFiles)

	mux := http.NewServeMux()
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return err
	}
	mux.Handle("GET /", http.FileServerFS(sub))
	mux.HandleFunc("GET /api/feeds", h.listFeeds)
	mux.HandleFunc("PUT /api/feeds", h.refreshFeeds)
	mux.HandleFunc("POST /api/feeds", h.addFeeds)
	mux.HandleFunc("POST /api/feeds/search", h.searchOnFeeds)

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.TelemetryMiddleware(mux),
	}
	return s.ListenAndServe()
}

package api

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"

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
	resp, err := h.feedAdapter.All(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "plain/text")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "error: %s", err)
		return
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

	flusher, ok := w.(http.Flusher)
	if !ok {
		var res []FeedSummary
		for _, f := range req.Feeds {
			feed, err := h.feedAdapter.Parse(ctx, f)
			if err != nil {
				log.WithError(err).Error("failed to encode feed summary")
				http.Error(w, fmt.Sprintf("failed to encode feed summary: %s", err), http.StatusInternalServerError)
				return
			}
			feedSummary := ToFeedSummary(feed)
			res = append(res, *feedSummary)
		}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.WithError(err).Error("failed to encode feed summary")
			http.Error(w, fmt.Sprintf("failed to encode feed summary: %s", err), http.StatusInternalServerError)
		}
		return
	}

	// Create a JSON encoder that writes directly to the response writer
	encoder := json.NewEncoder(w)

	defer func() {
		_ = r.Body.Close()
	}()

	log = log.WithExtraData("feeds_count", len(req.Feeds))

	log.Info("feeds retrieved")

	w.WriteHeader(http.StatusContinue)

	var res []FeedSummary
	for _, f := range req.Feeds {
		feed, err := h.feedAdapter.Parse(ctx, f)
		if err != nil {
			log.WithError(err).WithExtraData("feed_name", f).Error("failed to refresh feed")
			b, err := json.Marshal(FeedSummary{
				Title: "",
				URL:   f,
				Error: err.Error(),
			})
			if err != nil {
				log.WithError(err).Error("failed to encode feed summary")
			}
			_, _ = w.Write(b)
			flusher.Flush()
			continue
		}
		feedSummary := ToFeedSummary(feed)
		if err := encoder.Encode(feedSummary); err != nil {
			log.WithError(err).Error("failed to encode feed summary")
			_, _ = w.Write([]byte(
				"failed to encode feed summary: " + err.Error() + "\n" +
					"Please check the logs for more details.",
			))
		}
		//_, _ = w.Write([]byte("\n"))

		res = append(res, *feedSummary)
		flusher.Flush()
	}

	if err := h.notifier.Notify(ctx, "Feeds refreshed"); err != nil {
		log.WithError(err).Error("failed to notify")
	}

	log.Info("feeds refreshed")

	if err := encoder.Encode(&res); err != nil {
		log.WithError(err).Error("failed to encode feed summary")
		_, _ = w.Write([]byte(
			"failed to encode feed summary: " + err.Error() + "\n" +
				"Please check the logs for more details.",
		))
	}
	flusher.Flush()
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

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.TelemetryMiddleware(mux),
	}
	return s.ListenAndServe()
}

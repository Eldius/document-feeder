package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/initial-config-go/http/server"
	"github.com/eldius/initial-config-go/logs"
	"net/http"
)

type handler struct {
	feedAdapter *adapter.FeedAdapter
	notifier    adapter.Notifier
}

func (h *handler) listFeeds(w http.ResponseWriter, r *http.Request) {
	resp, err := h.feedAdapter.All(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "plain/text")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("error: %s", err)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("error: %s", err)))
		return
	}
}

func (h *handler) refreshFeeds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logs.NewLogger(ctx)
	log.Info("refreshing feeds")

	w.Header().Set("Content-Type", "application/json")

	all, err := h.feedAdapter.All(ctx)
	if err != nil {
		log.WithError(err).Error("failed to get feeds")
		if err := h.notifier.Notify(ctx, fmt.Sprintf("Failed to get feeds: %s", err)); err != nil {
			log.WithError(err).Error("failed to notify")
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf("error: %s", err)))
		return
	}

	log = log.WithExtraData("feeds_count", len(all))

	log.Info("feeds retrieved")

	w.WriteHeader(http.StatusContinue)
	for _, f := range all {
		if err := h.feedAdapter.RefreshFeed(ctx, f); err != nil {
			log.WithError(err).WithExtraData("feed_name", f.Title).Error("failed to refresh feed")
			_, _ = w.Write([]byte(fmt.Sprintf("failed to refresh feed %s: %s", f.Title, err)))
			continue
		}
		_, _ = w.Write([]byte(fmt.Sprintf("refreshed feed %s\n", f.Title)))
	}

	if err := h.notifier.Notify(ctx, "Feeds refreshed"); err != nil {
		log.WithError(err).Error("failed to notify")
	}

	log.Info("feeds refreshed")

	w.WriteHeader(http.StatusOK)
}

func (h *handler) addFeeds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logs.NewLogger(ctx)
	log.Info("refreshing feeds")

	// Set headers for streaming and JSON content
	w.Header().Set("Content-Type", "application/json")
	// Use chunked encoding if available (Go's net/http handles this implicitly with Flush)
	w.Header().Set("Transfer-Encoding", "chunked")

	var req AddFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("error: %s", err)))
		return
	}

	// Check if the ResponseWriter supports the Flusher interface
	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Println("ResponseWriter does not support Flusher interface")
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
			b, err := json.Marshal(map[string]any{
				"feed":  f,
				"error": err.Error(),
			})
			if err != nil {
				fmt.Printf("failed to encode feed summary: %s\n", err)
				log.WithError(err).Error("failed to encode feed summary")
			}
			_, _ = w.Write(b)
			flusher.Flush()
			continue
		}
		feedSummary := ToFeedSummary(feed)
		if err := encoder.Encode(feedSummary); err != nil {
			fmt.Printf("failed to encode feed summary: %s\n", err)
			log.WithError(err).Error("failed to encode feed summary")
			_, _ = w.Write([]byte(
				"failed to encode feed summary: " + err.Error() + "\n" +
					"Please check the logs for more details.",
			))
		}

		res = append(res, *feedSummary)
		fmt.Println(" -> flushing")
		flusher.Flush()
	}

	if err := h.notifier.Notify(ctx, "Feeds refreshed"); err != nil {
		log.WithError(err).Error("failed to notify")
	}

	log.Info("feeds refreshed")

	if err := encoder.Encode(&res); err != nil {
		fmt.Printf("failed to encode feed summary: %s\n", err)
		log.WithError(err).Error("failed to encode feed summary")
		_, _ = w.Write([]byte(
			"failed to encode feed summary: " + err.Error() + "\n" +
				"Please check the logs for more details.",
		))
	}
	flusher.Flush()
}

func newHandler(feedAdapter *adapter.FeedAdapter, notifier adapter.Notifier) *handler {
	return &handler{
		feedAdapter: feedAdapter,
		notifier:    notifier,
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

	h := newHandler(a, n)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/feeds", h.listFeeds)
	mux.HandleFunc("PUT /api/feeds", h.refreshFeeds)
	mux.HandleFunc("POST /api/feeds", h.addFeeds)

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.TelemetryMiddleware(mux),
	}
	return s.ListenAndServe()
}

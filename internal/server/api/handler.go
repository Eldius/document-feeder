package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/initial-config-go/http/server"
	"github.com/eldius/initial-config-go/logs"
	"github.com/eldius/initial-config-go/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	w.Header().Set("Content-Type", "application/json")
	go func(ctx context.Context) {
		spanCtx := trace.SpanContextFromContext(ctx)
		detachedCtx := trace.ContextWithSpanContext(context.Background(), spanCtx)

		detachedCtx, span := telemetry.NewSpan(detachedCtx, "refresh feeds", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(
			attribute.String("feeds", "all"),
		))
		defer span.End()

		log := logs.NewLogger(detachedCtx)
		log.Info("refreshing feeds")

		if err := h.feedAdapter.Refresh(detachedCtx); err != nil {
			log.WithError(err).Error("failed to refresh feeds")
			if err := h.notifier.Notify(detachedCtx, fmt.Sprintf("Failed to refresh feeds: %s", err)); err != nil {
				log.WithError(err).Error("failed to notify")
			}
			return
		}
		if err := h.notifier.Notify(detachedCtx, "Feeds refreshed"); err != nil {
			log.WithError(err).Error("failed to notify")
		}

		log.Info("feeds refreshed")
	}(r.Context())

	w.WriteHeader(http.StatusNoContent)
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

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server.TelemetryMiddleware(mux),
	}
	return s.ListenAndServe()
}

package api

import (
	"context"
	"time"
)

// detachedContext wraps a context but ignores cancellation and deadlines.
type detachedContext struct {
	context.Context
}

func (d detachedContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (d detachedContext) Done() <-chan struct{} {
	return nil
}

func (d detachedContext) Err() error {
	return nil
}

// Value delegates to the original context for request-scoped data (e.g., user ID, logger, trace ID).
func (d detachedContext) Value(key interface{}) interface{} {
	return d.Context.Value(key)
}

// Detach returns a context that shares values but not the cancellation signal.
func Detach(ctx context.Context) context.Context {
	return detachedContext{Context: ctx}
}

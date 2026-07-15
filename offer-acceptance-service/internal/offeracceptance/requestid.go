package offeracceptance

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
)

// RequestIDHeader carries a request ID across service boundaries so one
// value can be grepped across every service's logs for a single request.
const RequestIDHeader = "X-Request-Id"

type contextKey int

const requestIDContextKey contextKey = iota

func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// WithRequestID honors an inbound X-Request-Id header (propagated from an
// upstream service call) or mints a fresh one if this is the entry point,
// echoes it back on the response, and stores it in the request context.
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDContextKey, id)))
	})
}

func requestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDContextKey).(string)
	return id
}

// LoggerFrom returns base enriched with the request's requestId field, if
// the context carries one (it always will for requests that came through
// WithRequestID, but callers may pass a context that didn't).
func LoggerFrom(ctx context.Context, base *slog.Logger) *slog.Logger {
	if id := requestIDFromContext(ctx); id != "" {
		return base.With("requestId", id)
	}
	return base
}

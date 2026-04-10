package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type rateLimitEntry struct {
	count     int
	windowEnd time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*rateLimitEntry
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		entries: make(map[string]*rateLimitEntry),
		limit:   limit,
		window:  window,
	}
}

// Allow checks if a request from the given IP is allowed.
// Returns true if allowed, or false with the time until the window resets.
func (rl *rateLimiter) Allow(ip string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Lazy cleanup of stale entries
	for k, e := range rl.entries {
		if now.After(e.windowEnd) {
			delete(rl.entries, k)
		}
	}

	entry, exists := rl.entries[ip]
	if !exists || now.After(entry.windowEnd) {
		rl.entries[ip] = &rateLimitEntry{
			count:     1,
			windowEnd: now.Add(rl.window),
		}
		return true, 0
	}

	if entry.count >= rl.limit {
		retryAfter := time.Until(entry.windowEnd)
		return false, retryAfter
	}

	entry.count++
	return true, 0
}

type contextKey string

const requestIDKey contextKey = "request_id"

// RequestID extracts the request ID from a context.
func RequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// clientIP extracts the client's real IP address from the request,
// respecting X-Forwarded-For for proxied deployments.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// First IP in the chain is the original client
		if ip, _, ok := strings.Cut(xff, ","); ok {
			return strings.TrimSpace(ip)
		}
		return strings.TrimSpace(xff)
	}
	// Strip port from RemoteAddr
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// rateLimitMiddleware wraps a handler with per-IP rate limiting.
func rateLimitMiddleware(rl *rateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		allowed, retryAfter := rl.Allow(ip)
		if !allowed {
			slog.Warn("rate limited", "ip", ip, "retry_after_s", int(retryAfter.Seconds()), "request_id", RequestID(r.Context()))
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())+1))
			writeJSON(w, http.StatusTooManyRequests, errorResponse{
				Error: fmt.Sprintf("Too many requests. Please try again in %d seconds.", int(retryAfter.Seconds())+1),
			})
			return
		}
		next(w, r)
	}
}

// requestIDMiddleware assigns a unique request ID to each request.
// If the client sends X-Request-ID, that value is used.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// statusRecorder wraps ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs method, path, status, duration, and request ID.
func loggingMiddleware(slowThreshold time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", duration.Milliseconds(),
				"request_id", RequestID(r.Context()),
			}

			if duration > slowThreshold {
				slog.Warn("slow request", attrs...)
			} else {
				slog.Info("request", attrs...)
			}
		})
	}
}

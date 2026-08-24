package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	rateLimitRequests = 100
	rateLimitWindow   = time.Minute
)

type clientRate struct {
	windowStart time.Time
	requests    int
}

var rateLimitState = struct {
	sync.Mutex
	clients map[string]clientRate
}{
	clients: make(map[string]clientRate),
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := clientIP(r)
		now := time.Now()

		rateLimitState.Lock()
		state := rateLimitState.clients[client]
		if state.windowStart.IsZero() || now.Sub(state.windowStart) >= rateLimitWindow {
			state = clientRate{windowStart: now}
		}
		state.requests++
		rateLimitState.clients[client] = state
		allowed := state.requests <= rateLimitRequests
		remaining := rateLimitRequests - state.requests
		if remaining < 0 {
			remaining = 0
		}
		rateLimitState.Unlock()

		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", formatRateLimitValue(remaining))
		if !allowed {
			retryAfter := int(time.Until(state.windowStart.Add(rateLimitWindow)).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", formatRateLimitValue(retryAfter))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	if r.RemoteAddr == "" {
		return "unknown"
	}
	return r.RemoteAddr
}

func formatRateLimitValue(value int) string {
	return strconv.Itoa(value)
}

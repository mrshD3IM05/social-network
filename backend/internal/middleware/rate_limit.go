package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	// made it 1000 to not limit images reading and keep it global
	rateLimitRequests = 1000
	rateLimitWindow   = time.Minute
	// After this long without a request a client's window is forgotten, so the
	// map cannot grow without bound over the life of the process.
	rateLimitIdleTTL = 10 * time.Minute
)

var lastRateLimitSweep time.Time

// sweepRateLimitState drops windows nobody has touched recently.
// The caller must hold rateLimitState's lock.
func sweepRateLimitState(now time.Time) {
	if now.Sub(lastRateLimitSweep) < rateLimitWindow {
		return
	}
	lastRateLimitSweep = now
	for client, state := range rateLimitState.clients {
		if now.Sub(state.windowStart) > rateLimitIdleTTL {
			delete(rateLimitState.clients, client)
		}
	}
}

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
		sweepRateLimitState(now)
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

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rateLimitRequests))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		if !allowed {
			retryAfter := int(time.Until(state.windowStart.Add(rateLimitWindow)).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
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

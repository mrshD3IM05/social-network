package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type clientRate struct {
	windowStart time.Time
	requests    int
}

// limiter counts requests per client IP in fixed windows.
type limiter struct {
	sync.Mutex
	requests  int
	window    time.Duration
	clients   map[string]clientRate
	lastSweep time.Time
}

func newLimiter(requests int, window time.Duration) *limiter {
	return &limiter{requests: requests, window: window, clients: make(map[string]clientRate)}
}

// global: made it 1000 to not limit images reading
var globalLimiter = newLimiter(1000, time.Minute)

// auth: login and register, to slow down password guessing
var authLimiter = newLimiter(10, time.Minute)

func RateLimit(next http.Handler) http.Handler { return globalLimiter.middleware(next) }

func AuthRateLimit(next http.Handler) http.Handler { return authLimiter.middleware(next) }

func (l *limiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed, remaining, retryAfter := l.take(clientIP(r), time.Now())

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(l.requests))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *limiter) take(client string, now time.Time) (allowed bool, remaining, retryAfter int) {
	l.Lock()
	defer l.Unlock()

	// forget clients whose window is over, so the map does not grow forever
	if now.Sub(l.lastSweep) >= l.window {
		for key, state := range l.clients {
			if now.Sub(state.windowStart) >= l.window {
				delete(l.clients, key)
			}
		}
		l.lastSweep = now
	}

	state := l.clients[client]
	if state.windowStart.IsZero() || now.Sub(state.windowStart) >= l.window {
		state = clientRate{windowStart: now}
	}
	state.requests++
	l.clients[client] = state

	remaining = max(l.requests-state.requests, 0)
	retryAfter = max(int(state.windowStart.Add(l.window).Sub(now).Seconds()), 1)
	return state.requests <= l.requests, remaining, retryAfter
}

// clientIP is the address of the visitor. Behind Caddy every request comes
// from the proxy, so the real address is the last one Caddy added to
// X-Forwarded-For. That header is only read when the request comes from a
// private or loopback address (the proxy); anyone else could write it.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			parts := strings.Split(forwarded, ",")
			if last := strings.TrimSpace(parts[len(parts)-1]); net.ParseIP(last) != nil {
				return last
			}
		}
	}
	if host == "" {
		return "unknown"
	}
	return host
}

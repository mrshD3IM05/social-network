package websocket

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

// checkOrigin replaces the previous "accept everything" upgrader check.
// Same-origin handshakes are allowed, local development hosts are allowed, and
// anything else must be listed in ALLOWED_ORIGINS (comma separated).
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Non-browser clients do not send Origin; the session cookie is still required.
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if strings.EqualFold(parsed.Host, r.Host) {
		return true
	}
	switch parsed.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	for _, allowed := range strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",") {
		allowed = strings.TrimSpace(allowed)
		if allowed != "" && strings.EqualFold(allowed, origin) {
			return true
		}
	}
	return false
}

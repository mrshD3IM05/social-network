package middleware

import "net/http"

// SecurityHeaders is set on every API answer: no MIME sniffing, the API can
// never be shown inside a frame, and no referrer leaves the site.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		header.Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

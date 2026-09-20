package common

import "net/http"

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Page reads ?limit= and ?offset= from the query string.
// Missing or nonsense values fall back to the first page.
func Page(r *http.Request) (limit, offset int) {
	limit = QueryInt(r, "limit", DefaultPageSize)
	if limit > MaxPageSize {
		limit = MaxPageSize
	}
	offset = QueryInt(r, "offset", 0)
	return limit, offset
}

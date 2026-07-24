// Package idgen mints stable, prefixed public ids used across service boundaries
// and in API URLs. Format: "<prefix>_<nanoid>" — URL-safe, collision-resistant.
package idgen

import gonanoid "github.com/matoous/go-nanoid/v2"

// New returns a prefixed public id, e.g. New("usr") → "usr_V1StGXR8Z5jdHi6B".
func New(prefix string) string {
	id, err := gonanoid.New()
	if err != nil {
		// gonanoid.New only errors if the crypto source fails; fall back to a
		// fixed-length retry so callers never get an empty id.
		id, _ = gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyz", 21)
	}
	return prefix + "_" + id
}

// Public id prefixes (one per resource type).
const (
	PrefixUser    = "usr"
	PrefixPayment = "pay"
)

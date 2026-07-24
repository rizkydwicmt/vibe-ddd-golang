// Package hmac exposes constant-time comparison helpers for secret material
// (ops tokens, signatures) so equality checks don't leak timing oracles.
package hmac

import "crypto/subtle"

// Equal reports whether a and b are equal in constant time.
func Equal(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

package crypto

import (
	"crypto/sha256"
	"encoding/base64"
)

func Hash256(input string) []byte {
	h := sha256.New()
	h.Write([]byte(input))
	return h.Sum(nil)
}

func Hash32String(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	bytes := h.Sum(nil)[:24]
	return base64.URLEncoding.EncodeToString(bytes)
}

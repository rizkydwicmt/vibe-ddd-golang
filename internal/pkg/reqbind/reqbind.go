// Package reqbind decodes a handler's request body/query so handlers call a single
// binding seam instead of gin directly.
package reqbind

import (
	"github.com/gin-gonic/gin"
)

// Bind decodes the request body into out. An empty body is not an error.
func Bind(c *gin.Context, out any) error {
	return c.ShouldBindJSON(out)
}

// Query returns an optional request param. Returns "" when absent.
func Query(c *gin.Context, key string) string {
	return c.Query(key)
}

// BindQuery decodes query parameters into out (using `form` struct tags), applying the
// `validate:"..."` validation rules. An absent query is not an error.
func BindQuery(c *gin.Context, out any) error {
	return c.ShouldBindQuery(out)
}

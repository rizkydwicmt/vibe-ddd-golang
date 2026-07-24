package reqctx

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	keyRequestID        = "requestId"
	keyStartTime        = "start-time"
	keyRequestedVersion = "apiVersion.requested"
	keySelectedVersion  = "apiVersion.selected" // resolved version returned in debug.version
)

func SetRequestID(c *gin.Context, id string) { c.Set(keyRequestID, id) }
func RequestID(c *gin.Context) string        { return c.GetString(keyRequestID) }

type ctxKey struct{}

// WithRequestID returns a derived context carrying the requestId.
func WithRequestID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxKey{}, id)
}

// RequestIDFromContext extracts the requestId from a context.Context.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}

// SetStartTime / StartTime — request edge time for runtimeMs.
func SetStartTime(c *gin.Context, t time.Time) { c.Set(keyStartTime, t) }
func StartTime(c *gin.Context) time.Time {
	if v, ok := c.Get(keyStartTime); ok {
		if t, ok := v.(time.Time); ok {
			return t
		}
	}
	return time.Now()
}

// Requested/Selected API version.
func SetRequestedVersion(c *gin.Context, v string) { c.Set(keyRequestedVersion, v) }
func RequestedVersion(c *gin.Context) string       { return c.GetString(keyRequestedVersion) }
func SetSelectedVersion(c *gin.Context, v string)  { c.Set(keySelectedVersion, v) }
func SelectedVersion(c *gin.Context) string        { return c.GetString(keySelectedVersion) }

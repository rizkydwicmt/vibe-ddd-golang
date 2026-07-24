package middleware

import (
	"time"

	"vibe-ddd-golang/internal/pkg/reqctx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestInit stamps a server-generated requestId and the request start time
// onto the context at the edge before validation and handler execution.
//
// requestId is a sortable UUIDv7 with a `req_` prefix. The same value is echoed
// in the X-Request-Id response header for support/tracing.
func RequestInit() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := newRequestID()
		reqctx.SetRequestID(c, id)
		reqctx.SetStartTime(c, time.Now())
		c.Writer.Header().Set("X-Request-Id", id)
		// Propagate into the request context.Context so the logger and any
		// downstream (repos, workers) can read the id off ctx.
		c.Request = c.Request.WithContext(reqctx.WithRequestID(c.Request.Context(), id))
		c.Next()
	}
}

func newRequestID() string {
	v, err := uuid.NewV7()
	if err != nil {
		v = uuid.New() // v4 fallback; still unique, loses time-sortability
	}
	return "req_" + v.String()
}

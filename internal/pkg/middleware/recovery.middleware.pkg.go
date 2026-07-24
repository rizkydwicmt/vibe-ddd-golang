package middleware

import (
	"fmt"

	"vibe-ddd-golang/internal/common/enum"
	types "vibe-ddd-golang/internal/common/type"
	"vibe-ddd-golang/internal/pkg/logger"
	"vibe-ddd-golang/internal/pkg/reqctx"

	"github.com/gin-gonic/gin"
)

// Recovery converts a panic into a standard 500 envelope without crashing the
// process. The panic is always logged with a stack trace; the response uses the
// installed `send` so the response shape stays consistent.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			logger.ErrorWithStack("panic: %v path=%v requestId=%v",
				r, c.Request.URL.Path, reqctx.RequestID(c))

			Send(c, &types.Response{
				Code:    enum.CodeInternalError,
				Message: "Internal server error.",
				Err:     fmt.Errorf("panic: %v", r),
			})
		}()
		c.Next()
	}
}

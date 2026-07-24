package middleware

import (
	types "vibe-ddd-golang/internal/common/type"
	responsehelper "vibe-ddd-golang/internal/pkg/helper/response"

	"github.com/gin-gonic/gin"
)

func ResponseInit() gin.HandlerFunc {
	return func(c *gin.Context) {
		exposeRawError := gin.Mode() == gin.DebugMode

		c.Set("send", func(r *types.Response) {
			env := responsehelper.BuildResponse(c, r, exposeRawError)

			// Error-triggered debug logging (nothing on 2xx/3xx → cheap in prod):
			// >=400 logs metadata + req/resp payload + cause; 5xx adds a stack.
			responsehelper.LogResponseError(c, r.HTTPStatus, r.Code, r.Message, r.Data, r.Err)

			c.Abort()
			c.JSON(r.HTTPStatus, env)
		})

		c.Next()
	}
}

// Send is a convenience wrapper for handlers that prefer a direct call over
func Send(c *gin.Context, r *types.Response) {
	send := c.MustGet("send").(func(r *types.Response))
	send(r)
	return
}

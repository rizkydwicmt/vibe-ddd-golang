package ginctx

import "github.com/gin-gonic/gin"

// ContextValue returns nil when a Gin context key was not set.
func ContextValue(c *gin.Context, key string) any {
	v, ok := c.Get(key)
	if !ok {
		return nil
	}
	return v
}

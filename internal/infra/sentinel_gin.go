package infra

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SentinelIngressMiddleware rejects when ingress flow rule triggers.
func SentinelIngressMiddleware(enabled bool) gin.HandlerFunc {
	if !enabled {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		e, blockErr := IngressEntry()
		if blockErr != nil {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		defer e.Exit()
		c.Next()
	}
}

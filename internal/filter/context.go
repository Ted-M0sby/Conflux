package filter

import (
	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt/v5"
)

const ctxKeyFilter = "nexus_filter_ctx"

// Context carries per-request metadata for filters.
type Context struct {
	RequestID string
	Claims    jwt.MapClaims
}

// FromGin returns filter context, creating one if missing.
func FromGin(c *gin.Context) *Context {
	if v, ok := c.Get(ctxKeyFilter); ok {
		if fc, ok := v.(*Context); ok {
			return fc
		}
	}
	fc := &Context{}
	c.Set(ctxKeyFilter, fc)
	return fc
}

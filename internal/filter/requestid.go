package filter

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const headerRequestID = "X-Request-Id"

// RequestIDFilter ensures a request id exists.
type RequestIDFilter struct{}

func (RequestIDFilter) Name() string { return "request_id" }

func (RequestIDFilter) Order() int { return 0 }

func (RequestIDFilter) Handle(c *gin.Context, ctx *Context) bool {
	rid := c.GetHeader(headerRequestID)
	if rid == "" {
		rid = uuid.NewString()
		c.Request.Header.Set(headerRequestID, rid)
	}
	c.Writer.Header().Set(headerRequestID, rid)
	ctx.RequestID = rid
	return true
}

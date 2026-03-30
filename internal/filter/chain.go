package filter

import (
	"sort"

	"github.com/gin-gonic/gin"
)

// Filter is one step in the responsibility chain.
type Filter interface {
	Name() string
	Order() int
	Handle(c *gin.Context, ctx *Context) (proceed bool)
}

// Chain runs ordered filters.
type Chain struct {
	filters []Filter
}

// NewChain builds a chain with filters sorted by Order (ascending).
func NewChain(filters ...Filter) *Chain {
	cp := append([]Filter(nil), filters...)
	sort.SliceStable(cp, func(i, j int) bool {
		if cp[i].Order() != cp[j].Order() {
			return cp[i].Order() < cp[j].Order()
		}
		return cp[i].Name() < cp[j].Name()
	})
	return &Chain{filters: cp}
}

// Names returns registered filter names in execution order.
func (ch *Chain) Names() []string {
	out := make([]string, 0, len(ch.filters))
	for _, f := range ch.filters {
		out = append(out, f.Name())
	}
	return out
}

// Middleware adapts the chain to Gin.
func (ch *Chain) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := FromGin(c)
		for _, f := range ch.filters {
			if !f.Handle(c, ctx) {
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

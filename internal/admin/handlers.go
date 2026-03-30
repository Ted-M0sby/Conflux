package admin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"nexus/internal/filter"
	"nexus/internal/gateway/health"
	"nexus/internal/gateway/router"
)

// Register mounts read-only admin routes on rg (caller should apply auth middleware if needed).
func Register(rg *gin.RouterGroup, store *router.Store, chain *filter.Chain, hr *health.TargetsRunner) {
	rg.GET("/routes", func(c *gin.Context) {
		t := store.Get()
		if t == nil {
			c.JSON(http.StatusOK, gin.H{"routes": []any{}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"routes": t.Routes()})
	})
	rg.GET("/filters", func(c *gin.Context) {
		if chain == nil {
			c.JSON(http.StatusOK, gin.H{"filters": []string{}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"filters": chain.Names()})
	})
	rg.GET("/health", func(c *gin.Context) {
		t := store.Get()
		if t == nil || hr == nil {
			c.JSON(http.StatusOK, gin.H{"targets": gin.H{}})
			return
		}
		m := gin.H{}
		seen := map[string]struct{}{}
		for _, r := range t.Routes() {
			for _, u := range r.Targets {
				u = strings.TrimSpace(u)
				if u == "" {
					continue
				}
				if _, ok := seen[u]; ok {
					continue
				}
				seen[u] = struct{}{}
				m[u] = hr.Healthy(u)
			}
		}
		c.JSON(http.StatusOK, gin.H{"targets": m})
	})
}

func AdminAuthMiddleware(token string) gin.HandlerFunc {
	if strings.TrimSpace(token) == "" {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		if c.GetHeader("X-Admin-Token") != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

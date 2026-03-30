package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	addr := ":8080"
	if v := os.Getenv("MOCK_ADDR"); v != "" {
		addr = v
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.GET("/user/*p", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "mock-backend",
			"path":    c.Request.URL.Path,
			"ts":      time.Now().UTC().Format(time.RFC3339Nano),
		})
	})

	r.POST("/user/*p", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	if err := r.Run(addr); err != nil {
		panic(err)
	}
}

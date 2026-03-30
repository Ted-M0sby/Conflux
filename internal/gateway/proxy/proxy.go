package proxy

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"nexus/internal/gateway/router"
)

// Director mutates outbound req to hit target with optional path strip.
func buildDirector(route *router.Route, target *url.URL) func(*http.Request) {
	return func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		p := req.URL.Path
		if route.StripPrefix && strings.HasPrefix(p, route.PathPrefix) {
			p = strings.TrimPrefix(p, route.PathPrefix)
			if p == "" {
				p = "/"
			} else if !strings.HasPrefix(p, "/") {
				p = "/" + p
			}
		}
		req.URL.Path = p
		req.Host = target.Host
	}
}

// NewHandler returns a gin.HandlerFunc that reverse-proxies using balancer and table match.
func NewHandler(
	getTable func() *router.Table,
	balancer Balancer,
	setXRealIP bool,
	transport http.RoundTripper,
) gin.HandlerFunc {
	if transport == nil {
		transport = defaultTransport()
	}

	return func(c *gin.Context) {
		table := getTable()
		if table == nil {
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		route, ok := table.Match(c.Request.URL.Path)
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		targetURL, err := balancer.Pick(route)
		if err != nil {
			c.AbortWithStatus(http.StatusBadGateway)
			return
		}

		if setXRealIP {
			req := c.Request
			if rip := req.Header.Get("X-Forwarded-For"); rip != "" {
				c.Request.Header.Set("X-Forwarded-For", rip+", "+c.ClientIP())
			} else {
				c.Request.Header.Set("X-Forwarded-For", c.ClientIP())
			}
			c.Request.Header.Set("X-Real-IP", c.ClientIP())
		}

		rp := &httputil.ReverseProxy{
			Director:       buildDirector(route, targetURL),
			Transport:      transport,
			ModifyResponse: nil,
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				http.Error(w, "bad gateway", http.StatusBadGateway)
			},
		}
		rp.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

func defaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nexus/internal/admin"
	"nexus/internal/config"
	"nexus/internal/filter"
	"nexus/internal/gateway/health"
	"nexus/internal/gateway/proxy"
	"nexus/internal/gateway/router"
	"nexus/internal/infra"
)

// Options bundles runtime dependencies for the HTTP handler.
type Options struct {
	Config    *config.Config
	Store     *router.Store
	Health    *health.TargetsRunner
	Chain     *filter.Chain
	Transport http.RoundTripper
}

// NewEngine builds the Gin engine with admin, filters, Sentinel, and proxy fallback.
func NewEngine(opts Options) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	accessLog := filter.NewAccessLogger(opts.Config.LogAccessFile)
	r.Use(filter.AccessLogMiddleware(accessLog))

	if opts.Chain != nil {
		r.Use(opts.Chain.Middleware())
	}
	r.Use(infra.SentinelIngressMiddleware(opts.Config.Sentinel.Enabled))

	r.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	ag := r.Group(opts.Config.AdminPathPrefix, admin.AdminAuthMiddleware(opts.Config.AdminToken))
	admin.Register(ag, opts.Store, opts.Chain, opts.Health)

	balancer := selectBalancer(opts)
	rt := opts.Transport
	if rt == nil {
		rt = proxy.NewSentinelRoundTripper(opts.Config.Sentinel.Enabled, nil)
	}

	proxyHandler := proxy.NewHandler(
		func() *router.Table { return opts.Store.Get() },
		balancer,
		opts.Config.ProxyHeaderXReal,
		rt,
	)
	r.NoRoute(proxyHandler)
	return r
}

func selectBalancer(opts Options) proxy.Balancer {
	strategy := proxy.NormalizeStrategy(opts.Config.LBStrategy)
	var hf func(string) bool
	if opts.Health != nil {
		hf = opts.Health.Healthy
	}
	switch strategy {
	case "random":
		return proxy.NewRandomBalancer(hf)
	case "first":
		return proxy.FirstTargetBalancer{}
	default:
		return proxy.NewRoundRobinBalancer(hf)
	}
}

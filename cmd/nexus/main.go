package main

import (
	"context"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"nexus/internal/app"
	"nexus/internal/config"
	"nexus/internal/filter"
	"nexus/internal/gateway/health"
	"nexus/internal/gateway/proxy"
	"nexus/internal/gateway/router"
	"nexus/internal/infra"
	"nexus/internal/store"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cfg := config.Load()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if cfg.Sentinel.Enabled {
		if err := infra.InitSentinel(cfg.Sentinel); err != nil {
			log.Error("sentinel init failed", slog.Any("err", err))
			os.Exit(1)
		}
	}

	if _, err := store.Open(cfg.MySQL.DSN); err != nil {
		log.Error("mysql connect failed", slog.Any("err", err))
		os.Exit(1)
	}

	routeStore := &router.Store{}
	if err := loadInitialRoutes(cfg, routeStore, log); err != nil {
		log.Error("load routes failed", slog.Any("err", err))
		os.Exit(1)
	}

	var rdb *redis.Client
	if strings.TrimSpace(cfg.Redis.Addr) != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
	}

	limiter := buildLimiter(cfg, rdb)

	hc := health.NewTargetsRunner(
		time.Duration(cfg.Health.IntervalSec)*time.Second,
		time.Duration(cfg.Health.TimeoutSec)*time.Second,
		cfg.Health.HTTPPath,
		func() []string { return app.CollectTargets(routeStore.Get()) },
	)

	transport := proxy.NewSentinelRoundTripper(cfg.Sentinel.Enabled, nil)

	chain := filter.NewChain(
		filter.RequestIDFilter{},
		filter.JWTFilter{
			Secret:   []byte(cfg.JWTSecret),
			Required: cfg.JWTRequired,
			Skip:     cfg.JWTSKIPPrefixes,
		},
		filter.RateLimitFilter{
			Enable:  cfg.RateLimitEnable,
			Limiter: limiter,
		},
	)

	engine := app.NewEngine(app.Options{
		Config:    cfg,
		Store:     routeStore,
		Health:    hc,
		Chain:     chain,
		Transport: transport,
	})

	srv := &http.Server{Addr: cfg.ServerAddr, Handler: engine}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if strings.EqualFold(cfg.RoutesSource, "file") {
		watchCtx, cancelWatch := context.WithCancel(ctx)
		defer cancelWatch()
		if err := app.WatchYAML(watchCtx, cfg.RoutesFile, log, func(t *router.Table) error {
			routeStore.Set(t)
			return nil
		}); err != nil {
			log.Warn("routes file watch disabled", slog.Any("err", err))
		}
	} else if strings.EqualFold(cfg.RoutesSource, "nacos") {
		hosts := splitHosts(cfg.Nacos.ServerHosts)
		cli, err := router.BuildNacosConfigClient(
			cfg.Nacos.NamespaceID,
			cfg.Nacos.Username,
			cfg.Nacos.Password,
			cfg.Nacos.LogDir,
			cfg.Nacos.CacheDir,
			cfg.Nacos.NotLoadCache,
			hosts,
		)
		if err != nil {
			log.Error("nacos client failed", slog.Any("err", err))
			os.Exit(1)
		}
		loader := &router.NacosLoader{Client: cli}
		if err := loader.StartListen(cfg.Nacos.DataID, cfg.Nacos.Group, func(t *router.Table) error {
			routeStore.Set(t)
			return nil
		}); err != nil {
			log.Error("nacos listen failed", slog.Any("err", err))
			os.Exit(1)
		}
		log.Info("nacos route sync started")
	}

	go func() {
		log.Info("listening", slog.String("addr", cfg.ServerAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", slog.Any("err", err))
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	hc.Stop()
	if rdb != nil {
		_ = rdb.Close()
	}
	log.Info("bye")
}

func loadInitialRoutes(cfg *config.Config, store *router.Store, log *slog.Logger) error {
	if strings.EqualFold(cfg.RoutesSource, "nacos") {
		// Nacos loader will populate; keep local file optional bootstrap
		if _, err := os.Stat(cfg.RoutesFile); err == nil {
			t, err := router.LoadYAML(cfg.RoutesFile)
			if err == nil {
				store.Set(t)
				log.Info("bootstrapped routes from local file before nacos", slog.String("path", cfg.RoutesFile))
			}
		}
		return nil
	}
	t, err := router.LoadYAML(cfg.RoutesFile)
	if err != nil {
		return err
	}
	store.Set(t)
	return nil
}

func buildLimiter(cfg *config.Config, rdb *redis.Client) filter.Limiter {
	if !cfg.RateLimitEnable {
		return nil
	}
	window := time.Duration(cfg.RateLimitWindow) * time.Second
	if cfg.RedisRateLimit && rdb != nil {
		return filter.NewRedisSlidingWindow(rdb, "nexus:rl:", cfg.RateLimitRPS, window)
	}
	return filter.NewMemorySlidingWindow(cfg.RateLimitRPS, window, time.Second)
}

func splitHosts(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

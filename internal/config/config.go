package config

import (
	"os"
	"strconv"
	"strings"
)

// Config is application configuration (env-first for containers).
type Config struct {
	ServerAddr   string
	RoutesFile   string
	RoutesSource string // file | nacos

	Nacos NacosConfig

	JWTSecret       string
	JWTRequired     bool
	JWTSKIPPrefixes []string // e.g. /admin — comma-separated in env NEXUS_JWT_SKIP_PREFIXES
	RateLimitEnable bool
	RateLimitRPS    int // max requests per window per key
	RateLimitWindow int // seconds (sliding window length)

	AdminToken       string
	AdminPathPrefix  string
	LogAccessFile    string
	ProxyHeaderXReal bool

	MySQL   MySQLConfig
	Redis   RedisConfig
	Sentinel SentinelConfig
	Health  HealthConfig

	LBStrategy string // round_robin | random

	RedisRateLimit bool // use Redis for rate limit when true and Redis enabled
}

type NacosConfig struct {
	ServerHosts  string // comma-separated e.g. nacos:8848
	NamespaceID  string
	DataID       string
	Group        string
	Username     string
	Password     string
	LogDir       string
	CacheDir     string
	NotLoadCache bool
}

type MySQLConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type SentinelConfig struct {
	Enabled        bool
	FlowQPS        float64
	SlowRtMs       uint64
	SlowRatio      float64 // 0-1
	MinRequest     uint64
	StatIntervalMs uint32
}

type HealthConfig struct {
	IntervalSec int
	TimeoutSec  int
	HTTPPath    string // e.g. /health; empty => TCP dial only
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getenvFloat(key string, def float64) float64 {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func getenvUint64(key string, def uint64) uint64 {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func splitComma(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func getenvBool(key string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		ServerAddr:   getenv("NEXUS_ADDR", ":8080"),
		RoutesFile:   getenv("NEXUS_ROUTES_FILE", "configs/routes.yaml"),
		RoutesSource: strings.ToLower(getenv("NEXUS_ROUTES_SOURCE", "file")),

		Nacos: NacosConfig{
			ServerHosts:  getenv("NACOS_SERVER_HOSTS", "127.0.0.1:8848"),
			NamespaceID:  getenv("NACOS_NAMESPACE_ID", ""),
			DataID:       getenv("NACOS_ROUTES_DATA_ID", "nexus-routes.yaml"),
			Group:        getenv("NACOS_ROUTES_GROUP", "DEFAULT_GROUP"),
			Username:     getenv("NACOS_USERNAME", ""),
			Password:     getenv("NACOS_PASSWORD", ""),
			LogDir:       getenv("NACOS_LOG_DIR", "logs/nacos-sdk"),
			CacheDir:     getenv("NACOS_CACHE_DIR", "data/nacos-cache"),
			NotLoadCache: getenvBool("NACOS_NOT_LOAD_CACHE", true),
		},

		JWTSecret:       getenv("NEXUS_JWT_SECRET", "dev-secret-change-me"),
		JWTRequired:     getenvBool("NEXUS_JWT_REQUIRED", false),
		JWTSKIPPrefixes: splitComma(getenv("NEXUS_JWT_SKIP_PREFIXES", "/admin,/health")),
		RateLimitEnable: getenvBool("NEXUS_RATELIMIT_ENABLE", true),
		RateLimitRPS:    getenvInt("NEXUS_RATELIMIT_RPS", 100),
		RateLimitWindow: getenvInt("NEXUS_RATELIMIT_WINDOW_SEC", 10),

		AdminToken:       getenv("NEXUS_ADMIN_TOKEN", ""),
		AdminPathPrefix:  getenv("NEXUS_ADMIN_PREFIX", "/admin"),
		LogAccessFile:    getenv("NEXUS_ACCESS_LOG_FILE", ""),
		ProxyHeaderXReal: getenvBool("NEXUS_PROXY_X_REAL_IP", true),

		MySQL: MySQLConfig{
			DSN: getenv("NEXUS_MYSQL_DSN", ""),
		},
		Redis: RedisConfig{
			Addr:     getenv("NEXUS_REDIS_ADDR", ""),
			Password: getenv("NEXUS_REDIS_PASSWORD", ""),
			DB:       getenvInt("NEXUS_REDIS_DB", 0),
		},
		Sentinel: SentinelConfig{
			Enabled:        getenvBool("NEXUS_SENTINEL_ENABLE", true),
			FlowQPS:        getenvFloat("NEXUS_SENTINEL_FLOW_QPS", 5000),
			SlowRtMs:       getenvUint64("NEXUS_SENTINEL_SLOW_RT_MS", 800),
			SlowRatio:      getenvFloat("NEXUS_SENTINEL_SLOW_RATIO", 0.5),
			MinRequest:     getenvUint64("NEXUS_SENTINEL_MIN_REQUEST", 5),
			StatIntervalMs: uint32(getenvInt("NEXUS_SENTINEL_STAT_INTERVAL_MS", 1000)),
		},
		Health: HealthConfig{
			IntervalSec: getenvInt("NEXUS_HEALTH_INTERVAL_SEC", 10),
			TimeoutSec:  getenvInt("NEXUS_HEALTH_TIMEOUT_SEC", 2),
			HTTPPath:    getenv("NEXUS_HEALTH_HTTP_PATH", "/health"),
		},

		LBStrategy: getenv("NEXUS_LB", "round_robin"),

		RedisRateLimit: getenvBool("NEXUS_REDIS_RATELIMIT", false),
		// Redis for rate limit only when addr set and flag true
	}
}

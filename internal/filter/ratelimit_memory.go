package filter

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Limiter abstracts rate limiting (memory, Redis, Sentinel wrapper).
type Limiter interface {
	Allow(key string) bool
}

// MemorySlidingWindow is a fixed-bucket sliding window counter per key.
type MemorySlidingWindow struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	bucket   time.Duration
	state    map[string]*windowAgg
	lastClean time.Time
}

type windowAgg struct {
	mu      sync.Mutex
	buckets []int
	start   time.Time // aligned start of bucket 0
}

// NewMemorySlidingWindow creates a limiter: max `limit` events per `window`, bucket size `bucket` (e.g. 1s).
func NewMemorySlidingWindow(limit int, window, bucket time.Duration) *MemorySlidingWindow {
	if limit <= 0 {
		limit = 1
	}
	if window <= 0 {
		window = time.Second * 10
	}
	if bucket <= 0 {
		bucket = time.Second
	}
	return &MemorySlidingWindow{
		limit:   limit,
		window:  window,
		bucket:  bucket,
		state:   make(map[string]*windowAgg),
		lastClean: time.Now(),
	}
}

func (m *MemorySlidingWindow) Allow(key string) bool {
	now := time.Now()
	m.mu.Lock()
	if now.Sub(m.lastClean) > m.window*2 {
		m.cleanupLocked(now)
		m.lastClean = now
	}
	wa, ok := m.state[key]
	if !ok {
		wa = &windowAgg{buckets: make([]int, int(m.window/m.bucket)+1), start: now.Truncate(m.bucket)}
		m.state[key] = wa
	}
	m.mu.Unlock()

	wa.mu.Lock()
	defer wa.mu.Unlock()
	m.advance(wa, now)
	idx := int(now.Truncate(m.bucket).Sub(wa.start) / m.bucket)
	if idx < 0 || idx >= len(wa.buckets) {
		wa.buckets = make([]int, int(m.window/m.bucket)+1)
		wa.start = now.Truncate(m.bucket)
		idx = 0
	}
	sum := 0
	for _, c := range wa.buckets {
		sum += c
	}
	if sum >= m.limit {
		return false
	}
	wa.buckets[idx]++
	return true
}

func (m *MemorySlidingWindow) advance(wa *windowAgg, now time.Time) {
	need := int(m.window / m.bucket)
	if need <= 0 {
		need = 1
	}
	if len(wa.buckets) != need+1 {
		wa.buckets = make([]int, need+1)
		wa.start = now.Truncate(m.bucket)
		return
	}
	step := int(now.Truncate(m.bucket).Sub(wa.start) / m.bucket)
	if step <= 0 {
		return
	}
	if step >= len(wa.buckets) {
		for i := range wa.buckets {
			wa.buckets[i] = 0
		}
		wa.start = now.Truncate(m.bucket)
		return
	}
	// shift left
	copy(wa.buckets, wa.buckets[step:])
	for i := len(wa.buckets) - step; i < len(wa.buckets); i++ {
		wa.buckets[i] = 0
	}
	wa.start = wa.start.Add(time.Duration(step) * m.bucket)
}

func (m *MemorySlidingWindow) cleanupLocked(now time.Time) {
	for k, wa := range m.state {
		wa.mu.Lock()
		m.advance(wa, now)
		sum := 0
		for _, c := range wa.buckets {
			sum += c
		}
		wa.mu.Unlock()
		if sum == 0 {
			delete(m.state, k)
		}
	}
}

// RateLimitFilter applies Limiter per client IP + path.
type RateLimitFilter struct {
	Enable  bool
	Limiter Limiter
}

func (RateLimitFilter) Name() string { return "rate_limit" }

func (RateLimitFilter) Order() int { return 20 }

func (f RateLimitFilter) Handle(c *gin.Context, ctx *Context) bool {
	if !f.Enable || f.Limiter == nil {
		return true
	}
	key := c.ClientIP() + "|" + c.Request.URL.Path
	if !f.Limiter.Allow(key) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		return false
	}
	return true
}

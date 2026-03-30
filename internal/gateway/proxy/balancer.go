package proxy

import (
	"errors"
	"math/rand"
	"net/url"
	"strings"
	"sync/atomic"

	"nexus/internal/gateway/router"
)

var errNoHealthyTarget = errors.New("no healthy upstream target")

// Balancer picks upstream targets for a route.
type Balancer interface {
	Pick(route *router.Route) (*url.URL, error)
}

type healthFn func(target string) bool

// RoundRobinBalancer uses atomic counter over healthy targets.
type RoundRobinBalancer struct {
	health healthFn
	ctr    uint64
}

// NewRoundRobinBalancer returns a round-robin picker; health nil => all healthy.
func NewRoundRobinBalancer(health healthFn) *RoundRobinBalancer {
	return &RoundRobinBalancer{health: health}
}

func (b *RoundRobinBalancer) Pick(route *router.Route) (*url.URL, error) {
	targets := healthyTargets(route.Targets, b.health)
	if len(targets) == 0 {
		return nil, errNoHealthyTarget
	}
	i := atomic.AddUint64(&b.ctr, 1) - 1
	t := targets[int(i)%len(targets)]
	return url.Parse(t)
}

// RandomBalancer picks a random healthy target.
type RandomBalancer struct {
	health healthFn
}

// NewRandomBalancer returns a random picker.
func NewRandomBalancer(health healthFn) *RandomBalancer {
	return &RandomBalancer{health: health}
}

func (b *RandomBalancer) Pick(route *router.Route) (*url.URL, error) {
	targets := healthyTargets(route.Targets, b.health)
	if len(targets) == 0 {
		return nil, errNoHealthyTarget
	}
	t := targets[rand.Intn(len(targets))]
	return url.Parse(t)
}

// FirstTargetBalancer always uses targets[0] (phase-one style).
type FirstTargetBalancer struct{}

func (FirstTargetBalancer) Pick(route *router.Route) (*url.URL, error) {
	if len(route.Targets) == 0 {
		return nil, errNoHealthyTarget
	}
	return url.Parse(route.Targets[0])
}

func healthyTargets(all []string, health healthFn) []string {
	if health == nil {
		return append([]string(nil), all...)
	}
	var out []string
	for _, t := range all {
		if health(t) {
			out = append(out, t)
		}
	}
	return out
}

// NormalizeStrategy returns round_robin | random | first.
func NormalizeStrategy(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "random":
		return "random"
	case "first", "":
		return "first"
	default:
		return "round_robin"
	}
}

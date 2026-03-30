package proxy

import (
	"testing"

	"nexus/internal/gateway/router"
)

func TestRoundRobinSkipsUnhealthy(t *testing.T) {
	r := &router.Route{Targets: []string{"http://a", "http://b"}}
	b := NewRoundRobinBalancer(func(s string) bool {
		return s == "http://b"
	})
	u, err := b.Pick(r)
	if err != nil {
		t.Fatal(err)
	}
	if u.String() != "http://b" {
		t.Fatalf("unexpected %s", u)
	}
}

func TestRandomAndFirst(t *testing.T) {
	rt := &router.Route{Targets: []string{"http://only"}}
	_, err := (FirstTargetBalancer{}).Pick(rt)
	if err != nil {
		t.Fatal(err)
	}
	rb := NewRandomBalancer(nil)
	_, err = rb.Pick(rt)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeStrategy(t *testing.T) {
	if NormalizeStrategy("") != "first" || NormalizeStrategy("random") != "random" || NormalizeStrategy("round_robin") != "round_robin" {
		t.Fatal("normalize strategy mismatch")
	}
}

func TestNoHealthyTargets(t *testing.T) {
	r := &router.Route{Targets: []string{"http://dead"}}
	b := NewRoundRobinBalancer(func(string) bool { return false })
	_, err := b.Pick(r)
	if err != errNoHealthyTarget {
		t.Fatalf("expected errNoHealthyTarget, got %v", err)
	}
}

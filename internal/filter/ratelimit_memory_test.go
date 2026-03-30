package filter

import (
	"testing"
	"time"
)

func TestMemorySlidingWindow_Burst(t *testing.T) {
	m := NewMemorySlidingWindow(3, 2*time.Second, time.Second)
	key := "k1|/p"
	if !m.Allow(key) || !m.Allow(key) || !m.Allow(key) {
		t.Fatal("expected first 3 allowed")
	}
	if m.Allow(key) {
		t.Fatal("expected 4th blocked")
	}
}

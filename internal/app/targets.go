package app

import (
	"strings"

	"nexus/internal/gateway/router"
)

// CollectTargets returns unique upstream URLs from the routing table.
func CollectTargets(t *router.Table) []string {
	if t == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
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
			out = append(out, u)
		}
	}
	return out
}

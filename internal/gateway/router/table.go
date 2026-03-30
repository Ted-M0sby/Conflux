package router

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// Table holds immutable route rules for lock-free reads via atomic swap.
type Table struct {
	routes []Route // sorted: higher priority first, then longer PathPrefix
}

// NewTable validates and builds a routing table from raw routes.
func NewTable(routes []Route) (*Table, error) {
	cp := make([]Route, len(routes))
	copy(cp, routes)
	out := make([]Route, 0, len(cp))
	for _, r := range cp {
		if r.ID == "" || r.PathPrefix == "" || len(r.Targets) == 0 {
			continue
		}
		for _, t := range r.Targets {
			u, err := url.Parse(t)
			if err != nil || u.Scheme == "" || u.Host == "" {
				return nil, fmt.Errorf("invalid target %q", t)
			}
		}
		p := r.PathPrefix
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		r.PathPrefix = strings.TrimSuffix(p, "/")
		out = append(out, r)
	}
	if len(out) == 0 {
		return &Table{routes: nil}, nil
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority != out[j].Priority {
			return out[i].Priority > out[j].Priority
		}
		return len(out[i].PathPrefix) > len(out[j].PathPrefix)
	})
	return &Table{routes: out}, nil
}

// Match returns the first matching route for request path.
func (t *Table) Match(path string) (*Route, bool) {
	if t == nil {
		return nil, false
	}
	p := path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = strings.TrimSuffix(p, "/")
	if p == "" {
		p = "/"
	}
	for i := range t.routes {
		r := &t.routes[i]
		rp := r.PathPrefix
		if p == rp || strings.HasPrefix(p, rp+"/") {
			return r, true
		}
	}
	return nil, false
}

// Routes returns a shallow copy for admin / debugging.
func (t *Table) Routes() []Route {
	if t == nil {
		return nil
	}
	out := make([]Route, len(t.routes))
	copy(out, t.routes)
	return out
}

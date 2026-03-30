package proxy

import (
	"net/http"

	"nexus/internal/infra"
)

// NewSentinelRoundTripper wraps base transport with Sentinel outbound stats.
func NewSentinelRoundTripper(enabled bool, base http.RoundTripper) http.RoundTripper {
	if !enabled {
		return base
	}
	if base == nil {
		base = defaultTransport()
	}
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		e, blockErr := infra.UpstreamEntry()
		if blockErr != nil {
			return nil, blockErr
		}
		defer e.Exit()
		return base.RoundTrip(r)
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

package health

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TargetsRunner periodically probes all URLs returned by getURLs.
type TargetsRunner struct {
	interval time.Duration
	timeout  time.Duration
	httpPath string
	client   *http.Client
	getURLs  func() []string

	mu     sync.RWMutex
	status map[string]bool
	cancel context.CancelFunc
}

// NewTargetsRunner starts background health checks.
func NewTargetsRunner(interval, timeout time.Duration, httpPath string, getURLs func() []string) *TargetsRunner {
	tr := &TargetsRunner{
		interval: interval,
		timeout:  timeout,
		httpPath: strings.TrimSpace(httpPath),
		client: &http.Client{
			Timeout: timeout,
		},
		getURLs: getURLs,
		status:  make(map[string]bool),
	}
	ctx, cancel := context.WithCancel(context.Background())
	tr.cancel = cancel
	go tr.loop(ctx)
	return tr
}

func (tr *TargetsRunner) loop(ctx context.Context) {
	tick := time.NewTicker(tr.interval)
	defer tick.Stop()
	tr.runOnce()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			tr.runOnce()
		}
	}
}

func (tr *TargetsRunner) runOnce() {
	urls := tr.getURLs()
	next := make(map[string]bool, len(urls))
	for _, raw := range urls {
		next[raw] = tr.probe(raw)
	}
	tr.mu.Lock()
	tr.status = next
	tr.mu.Unlock()
}

// Stop stops the background loop.
func (tr *TargetsRunner) Stop() {
	if tr.cancel != nil {
		tr.cancel()
	}
}

// Healthy returns the last probe result. Unknown targets default to true.
func (tr *TargetsRunner) Healthy(target string) bool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	if v, ok := tr.status[target]; ok {
		return v
	}
	return true
}

func (tr *TargetsRunner) probe(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := u.Host
	if host == "" {
		return false
	}
	if tr.httpPath != "" {
		checkURL := *u
		checkURL.Path = tr.httpPath
		req, err := http.NewRequest(http.MethodGet, checkURL.String(), nil)
		if err != nil {
			return false
		}
		ctx, cancel := context.WithTimeout(context.Background(), tr.timeout)
		defer cancel()
		req = req.WithContext(ctx)
		resp, err := tr.client.Do(req)
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode < 500
	}
	d := net.Dialer{Timeout: tr.timeout}
	conn, err := d.Dial("tcp", host)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

package infra

import (
	"fmt"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"

	"nexus/internal/config"
)

const (
	ResourceIngress  = "gateway:ingress"
	ResourceUpstream = "gateway:upstream"
)

// InitSentinel configures default SDK and loads flow / circuit breaker rules.
func InitSentinel(cfg config.SentinelConfig) error {
	if !cfg.Enabled {
		return nil
	}
	if err := sentinel.InitDefault(); err != nil {
		return err
	}
	stat := cfg.StatIntervalMs
	if stat == 0 {
		stat = 1000
	}
	qps := cfg.FlowQPS
	if qps <= 0 {
		qps = 1000
	}
	if _, err := flow.LoadRules([]*flow.Rule{{
		Resource:               ResourceIngress,
		Threshold:              qps,
		TokenCalculateStrategy: flow.Direct,
		ControlBehavior:        flow.Reject,
		StatIntervalInMs:       uint32(stat),
	}}); err != nil {
		return fmt.Errorf("flow rules: %w", err)
	}
	if _, err := circuitbreaker.LoadRules([]*circuitbreaker.Rule{{
		Resource:         ResourceUpstream,
		Strategy:         circuitbreaker.SlowRequestRatio,
		RetryTimeoutMs:   3000,
		MinRequestAmount: cfg.MinRequest,
		StatIntervalMs:   stat,
		Threshold:        cfg.SlowRatio,
		MaxAllowedRtMs:   cfg.SlowRtMs,
	}}); err != nil {
		return fmt.Errorf("circuit breaker rules: %w", err)
	}
	return nil
}

// IngressEntry executes ingress Sentinel slot around gateway handling.
func IngressEntry() (*base.SentinelEntry, *base.BlockError) {
	return sentinel.Entry(ResourceIngress, sentinel.WithTrafficType(base.Inbound))
}

// UpstreamEntry tracks outbound backend RT for circuit breaking.
func UpstreamEntry() (*base.SentinelEntry, *base.BlockError) {
	return sentinel.Entry(ResourceUpstream, sentinel.WithTrafficType(base.Outbound))
}

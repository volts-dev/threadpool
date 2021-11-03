package threadpool

import "time"

// Option represents the optional function.
type (
	Option  func(opts *TConfig)
	TConfig struct {
		ExpiryDuration time.Duration
	}
)

func newConfig(opts ...Option) *TConfig {
	cfg := &TConfig{}

	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

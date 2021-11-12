package threadpool

import (
	"errors"
	"runtime"
	"time"
)

// Option represents the optional function.
type (
	Option  func(cfg *TConfig)
	TConfig struct {
		ExpiryDuration time.Duration
		MinThread      int
		MaxThread      int
		MaxTasks       int
		Capacity       int32
	}
)

const (
	RUNNING = iota // RUNNING represents that the pool is Actived.
	STOPPED
	CLOSED // CLOSED represents that the pool is closed.
)

var (
	// errQueueIsFull will be returned when the worker queue is full.
	errQueueIsFull = errors.New("the queue is full")
	// errQueueIsReleased will be returned when trying to insert item to a released worker queue.
	errQueueIsReleased = errors.New("the queue length is zero")
	defaultPool        = NewPool()
)

func newConfig(opts ...Option) *TConfig {
	cores := runtime.GOMAXPROCS(runtime.NumCPU())
	cfg := &TConfig{
		ExpiryDuration: 600 * time.Second,
		MinThread:      1,
		MaxThread:      cores * 2,
		MaxTasks:       cores * 8,
	}

	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// WithExpiryDuration sets up the interval time of cleaning up goroutines.
func WithExpiryDuration(expiryDuration time.Duration) Option {
	return func(cfg *TConfig) {
		cfg.ExpiryDuration = expiryDuration
	}
}

func MaxThread(cnt int) Option {
	return func(cfg *TConfig) {
		cfg.MaxThread = cnt
		cfg.MaxTasks = cnt * 3
	}
}

func MaxTasks(cnt int) Option {
	return func(cfg *TConfig) {
		cfg.MaxTasks = cnt
	}
}

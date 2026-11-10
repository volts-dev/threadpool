package threadpool

import (
	"errors"
	"runtime"
	"time"
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
)

// Option represents the optional function.
type (
	Option  func(cfg *TConfig)
	TConfig struct {
		ExpiryDuration time.Duration
		MinThread      int32
		MaxThread      int32
		Capacity       int32
		Delay          time.Duration
	}
)

func newConfig(opts ...Option) *TConfig {
	cores := int32(runtime.GOMAXPROCS(runtime.NumCPU()))
	cfg := &TConfig{
		ExpiryDuration: 3 * time.Second,
		Delay:          300 * time.Millisecond,
		MinThread:      1,
		MaxThread:      cores * 2,
		Capacity:       cores * 10,
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

// 最大处理线程
func WithMaxThread(cnt int) Option {
	return func(cfg *TConfig) {
		cfg.MaxThread = int32(cnt)
		cfg.Capacity = int32(cnt) * 3
	}
}

// 任务容量
func WithCapacity(cnt int) Option {
	return func(cfg *TConfig) {
		cfg.Capacity = int32(cnt)
	}
}

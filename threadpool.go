package threadpool

import (
	"sync"
	"time"
)

type (
	thread struct {
		// pool who owns this worker.
		pool *TThreadPool

		// task is a job should be done.
		task chan func()

		// recycleTime will be updated when putting a worker back into queue.
		recycleTime time.Time
	}

	ITask interface {
	}

	IController interface {
		Execute(ITask) bool
	}

	TThreadPool struct {
		config *TConfig

		task        chan func()
		threadCache sync.Pool
	}
)

func NewPool(opts ...Option) *TThreadPool {
	cfg := newConfig(opts...)

	pool := &TThreadPool{
		config: cfg,
	}

	pool.threadCache.New = func() interface{} {
		return pool.newThread()
	}

	// manager thread
	go func(p *TThreadPool) {
		heartbeat := time.NewTicker(p.config.ExpiryDuration)
		defer heartbeat.Stop()

		for range heartbeat.C {
			if p.IsClosed() {
				break
			}
		}
	}(pool)

	return pool
}

func (self *TThreadPool) newThread() *thread {
	return &thread{
		pool: self,
	}
}

func (self *TThreadPool) Active(sw bool) {

}

func (self *TThreadPool) Run() {

}

func (self *TThreadPool) Stop() {

}

func (self *TThreadPool) Release() {

}

package threadpool

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/volts-dev/utils"
)

type (
	TThreadPool struct {
		cond        *sync.Cond // cond for waiting to get a idle worker.
		config      *TConfig
		threadCount int32       // num of thread are running
		taskCount   int32       //
		state       int32       // state is used to notice the pool to closed itself.
		tasks       chan func() // 任务列队
		threads     *threadQueue
	}
)

func NewPool(opts ...Option) *TThreadPool {
	cfg := newConfig(opts...)

	pool := &TThreadPool{
		state:  STOPPED, // 初始状态为暂停状态不处理任务 必须调用Active(true)
		config: cfg,
		cond:   sync.NewCond(&sync.Mutex{}),
		tasks:  make(chan func(), cfg.Capacity),
	}
	pool.threads = newThreadQueue(pool)

	// manager thread
	go func(p *TThreadPool) {
		heartbeat := time.NewTicker(p.config.ExpiryDuration)
		defer heartbeat.Stop()

		for range heartbeat.C {
			if atomic.LoadInt32(&p.state) == CLOSED {
				break
			}

			// 确保最小线程
			if p.threadCount < p.config.MinThread {
				num := p.config.MinThread - p.threadCount
				p.AddThread(int(num))
			}

			// 弹性添加线程
			if p.taskCount > 0 && p.threadCount < p.config.MaxThread {
				if p.taskCount > p.config.MaxThread { // 任务量大于线程池容量 直接开最大
					p.AddThread(int(p.config.MaxThread - p.threadCount))
				} else {
					num := utils.MaxInt(int(p.taskCount-p.threadCount), 0)
					p.AddThread(num)
				}
			}
		}
	}(pool)

	pool.AddThread(int(cfg.MinThread))
	pool.Active(true)
	return pool
}

func (self *TThreadPool) Config(opts ...Option) *TConfig {
	for _, opt := range opts {
		opt(self.config)
	}

	return self.config
}

// clear all task
func (self *TThreadPool) Clear() {
	close(self.tasks)
	self.tasks = make(chan func(), self.config.Capacity)
}

func (self *TThreadPool) Close() {
	atomic.StoreInt32(&self.state, CLOSED)
	close(self.tasks)
}

// Cap returns the capacity of this pool.
func (self *TThreadPool) Cap() int {
	return int(self.config.Capacity)
}

// 设置激活状态
func (self *TThreadPool) Active(sw ...bool) bool {
	if sw != nil {
		if sw[0] {
			self.cond.Broadcast()
			atomic.StoreInt32(&self.state, RUNNING)
			return true
		} else {
			atomic.StoreInt32(&self.state, STOPPED)
			return false
		}
	}

	return int(atomic.LoadInt32(&self.state)) == RUNNING
}

func (self *TThreadPool) ThreadCount() int {
	return int(atomic.LoadInt32(&self.threadCount))
}

func (self *TThreadPool) IdleThreadCount() int {
	return int(atomic.LoadInt32(&self.threadCount))
}

func (self *TThreadPool) TaskCount() int {
	return int(atomic.LoadInt32(&self.taskCount))
}

// Submit submits a task to this pool.
func (self *TThreadPool) AddTask(task func()) error {
	self.tasks <- task
	self.incTask()
	return nil
}

// get a thread
func (self *TThreadPool) AddThread(num ...int) {
	if num != nil {
		for i := 0; i < num[0]; i++ {
			t := self.threads.Get()
			t.Start()
		}
	} else {
		t := self.threads.Get()
		t.Start()
	}
}

// release a thread to ide
func (self *TThreadPool) revertThread(t *thread) bool {
	if capacity := self.Cap(); (capacity > 0 && int(atomic.LoadInt32(&self.threadCount)) > capacity) || self.state == CLOSED {
		return false
	}
	t.recycleTime = time.Now()
	return self.threads.Put(t) == nil
}

// incRunning increases the number of the currently running goroutines.
func (self *TThreadPool) incThread() {
	atomic.AddInt32(&self.threadCount, 1)
}

// decRunning decreases the number of the currently running goroutines.
func (self *TThreadPool) decThread() {
	atomic.AddInt32(&self.threadCount, -1)
}

func (self *TThreadPool) incTask() {
	atomic.AddInt32(&self.taskCount, 1)
}

// decRunning decreases the number of the currently running goroutines.
func (self *TThreadPool) decTask() {
	atomic.AddInt32(&self.taskCount, -1)
}

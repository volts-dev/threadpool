package threadpool

import (
	"sync"
	"sync/atomic"
	"time"
)

type (
	TThreadPool struct {
		cond   *sync.Cond // cond for waiting to get a idle worker.
		config *TConfig
		//capacity    int32
		threadCount int32
		taskCount   int32
		state       int32 // state is used to notice the pool to closed itself.
		threads     *threadQueue
		tasks       chan func() // 任务列队
		threadCache sync.Pool
	}
)

func NewPool(opts ...Option) *TThreadPool {
	cfg := newConfig(opts...)

	pool := &TThreadPool{
		state:  STOPPED, // 初始状态为暂停状态不处理任务 必须调用Active(true)
		config: cfg,
		cond:   sync.NewCond(&sync.Mutex{}),
		tasks:  make(chan func(), cfg.MaxTasks),
	}
	pool.threads = newThreadQueue(pool)
	pool.threadCache.New = func() interface{} {
		return NewThread(pool)
	}
	pool.Active(true)

	// manager thread
	go func(p *TThreadPool) {
		heartbeat := time.NewTicker(p.config.ExpiryDuration)
		defer heartbeat.Stop()

		for range heartbeat.C {
			if atomic.LoadInt32(&p.state) == CLOSED {
				break
			}

		}
	}(pool)

	pool.AddThread(cfg.MaxThread)
	return pool
}

func (self *TThreadPool) Init(opts ...Option) {
	for _, opt := range opts {
		opt(self.config)
	}
}

// clear all task
func (self *TThreadPool) Clear() {
	close(self.tasks)
	self.tasks = make(chan func(), self.config.MaxTasks)
}

func (self *TThreadPool) Close() {
	atomic.StoreInt32(&self.state, CLOSED)
	close(self.tasks)
}

// Cap returns the capacity of this pool.
func (p *TThreadPool) Cap() int {
	return len(p.tasks)
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
			t := self.threadCache.Get().(*thread)
			t.Start()
			self.threads.Put(t)
			self.incThread()
		}
	} else {
		t := self.threadCache.Get().(*thread)
		t.Start()
		self.threads.Put(t)
		self.incThread()
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

func Init(opts ...Option) {
	defaultPool.Init(opts...)
}

func Cap() int {
	return defaultPool.Cap()
}

func Close() {
	defaultPool.Close()
}

func Active(sw ...bool) bool {
	return defaultPool.Active(sw...)
}

func ThreadCount() int {
	return defaultPool.ThreadCount()
}

func TaskCount() int {
	return defaultPool.TaskCount()
}

func AddTask(task func()) error {
	return defaultPool.AddTask(task)
}

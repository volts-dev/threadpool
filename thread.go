package threadpool

import (
	"sync"
	"sync/atomic"
	"time"
)

// goWorker is the actual executor who runs the tasks,
// it starts a goroutine that accepts tasks and
// performs function calls.
type (
	thread struct {
		// pool who owns this worker.
		pool *TThreadPool

		// task is a job should be done.
		task   chan func()
		signal chan int32
		// recycleTime will be updated when putting a worker back into queue.
		recycleTime time.Time
	}

	threadQueue struct {
		sync.RWMutex
		pool    *TThreadPool
		threads []*thread
		expiry  []*thread

		size int
	}
)

func NewThread(pool *TThreadPool) *thread {
	t := &thread{
		pool: pool,
		task: make(chan func()),
	}

	return t
}

func (self *thread) Run() {
	go func() {
		pool := self.pool

		defer func(pool *TThreadPool) {
			pool.threadCache.Put(self) // 回收为缓存
			pool.decThread()
		}(pool)

		cond := self.pool.cond
		for {
			// 待工开关
			cond.L.Lock()
			cond.Wait()
			cond.L.Unlock()
			if !pool.Active() {
				continue
			}

			// 检测线程数
			if int32(atomic.LoadInt32(&pool.threadCount)) > int32(pool.config.MaxThread) {
				return // 释放线程
			}

			for {
				// 这个方法会在 600 秒钟向 timeout 管道写入数据 防止死线程
				timeout := time.After(pool.config.ExpiryDuration)

				select {
				case task := <-pool.tasks:
					if task == nil {
						return // 释放线程
					}
					task()
					pool.decTask()
					if !pool.Active() {
						goto excape
					}
				case s := <-self.signal:
					if s > 0 {
						return
					}
				case <-timeout:
					// 过期
					return
				}
			}
		excape:
		}

	}()
}

func newThreadQueue(pool *TThreadPool) *threadQueue {
	return &threadQueue{
		pool:    pool,
		threads: make([]*thread, 0),
	}
}

func (self *threadQueue) Count() int {
	return self.size
}

func (self *threadQueue) Put(thread ...*thread) error {
	defer self.RUnlock()
	self.RLock()

	if self.size == 0 {
		return errQueueIsReleased
	}

	if self.size == self.pool.config.MaxThread {
		return errQueueIsFull
	}

	self.threads = append(self.threads, thread...)
	self.size = len(self.threads)
	return nil
}

func (self *threadQueue) Get() *thread {
	defer self.RUnlock()
	self.RLock()

	cnt := self.size
	if cnt == 0 {
		return nil
	}

	t := self.threads[0]
	self.threads = self.threads[1:]
	self.size = len(self.threads)
	return t
}

func (self *threadQueue) Clear() {
	if self.size == 0 {
		return
	}

	for {
		t := self.Get()
		if t == nil {
			return
		}

		t.task <- nil
	}
}

func (self *threadQueue) retrieveExpiry(duration time.Duration) []*thread {
	cnt := self.size
	if cnt == 0 {
		return nil
	}

	return nil
}

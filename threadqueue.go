package threadpool

import (
	"sync"
	"time"
)

type (
	threadQueue struct {
		sync.RWMutex
		pool        *TThreadPool
		expiry      []*thread
		threads     []*thread
		threadCache sync.Pool
		size        int32 /// num of thteads are idle and ready
	}
)

func newThreadQueue(pool *TThreadPool) *threadQueue {
	q := &threadQueue{
		pool:    pool,
		threads: make([]*thread, 0),
	}
	q.threadCache.New = func() interface{} {
		return NewThread(pool)
	}
	return q
}

func (self *threadQueue) Count() int {
	return int(self.size)
}

func (self *threadQueue) Put(thread ...*thread) error {
	defer self.RUnlock()
	self.RLock()

	//if self.size == 0 {
	//	return errQueueIsReleased
	//}

	if self.size == self.pool.config.MaxThread {
		for _, t := range thread {
			if t != nil {
				self.threadCache.Put(t)
			}
		}
		return errQueueIsFull
	}

	self.threads = append(self.threads, thread...)
	self.size = int32(len(self.threads))
	return nil
}

func (self *threadQueue) Get() *thread {
	defer self.RUnlock()
	self.RLock()

	var t *thread
	cnt := self.size
	if cnt == 0 {
		t = self.threadCache.Get().(*thread)
	} else {
		t = self.threads[0]
		self.threads = self.threads[1:]
	}

	self.size = int32(len(self.threads))
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

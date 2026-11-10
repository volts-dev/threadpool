package threadpool

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
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
)

func Goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func NewThread(pool *TThreadPool) *thread {
	t := &thread{
		pool: pool,
		task: make(chan func()),
	}

	return t
}

func (self *thread) Start() {
	go func() {
		pool := self.pool
		pool.incThread()
		defer func(pool *TThreadPool) {
			pool.threads.Put(self) // 回收为缓存
			pool.decThread()
		}(pool)

		cond := self.pool.cond
		for {

			if !pool.Active() {
				// 待工开关
				cond.L.Lock()
				cond.Wait()
				cond.L.Unlock()
				//	continue
			}

			// 检测线程数
			if (atomic.LoadInt32(&pool.threadCount)) > int32(pool.config.MaxThread) {
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

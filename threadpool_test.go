package threadpool

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	RunTimes           = 1000000
	BenchParam         = 10
	BenchAntsSize      = 200000
	DefaultExpiredTime = 10 * time.Second
)

const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
	// GiB // 1073741824
	// TiB // 1099511627776             (超过了int32的范围)
	// PiB // 1125899906842624
	// EiB // 1152921504606846976
	// ZiB // 1180591620717411303424    (超过了int64的范围)
	// YiB // 1208925819614629174706176
)

const (
	Param    = 100
	AntsSize = 1000
	TestSize = 10000
	n        = 100000
)

var curMem uint64

func demoFunc() {
	time.Sleep(time.Duration(BenchParam) * time.Millisecond)
	//fmt.Print("1")
}

func TestNoPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			demoFunc()
			wg.Done()
		}()
	}

	wg.Wait()
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

func TestPool(t *testing.T) {
	pool := NewPool()
	defer pool.Close()

	pool.Config(
		WithMaxThread(1000),
		WithCapacity(5000),
	)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = pool.AddTask(func() {
			demoFunc()
			wg.Done()
		})
	}
	wg.Wait()

	t.Logf("pool, capacity:%d", pool.Cap())
	t.Logf("pool, running thread number:%d", pool.ThreadCount())
	t.Logf("pool, task number:%d", pool.TaskCount())

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)

}

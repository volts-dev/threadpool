package threadpool

import (
	"sync"
	"testing"
)

func BenchmarkGoroutines(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			go func() {
				demoFunc()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkPool(b *testing.B) {
	var wg sync.WaitGroup
	p := NewPool(
		MaxThread(BenchAntsSize),
		WithExpiryDuration(DefaultExpiredTime),
	)
	defer p.Close()
	p.Active(true)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			_ = p.AddTask(func() {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
	b.StopTimer()
}

func BenchmarkAntsPoolThroughput(b *testing.B) {
	p := NewPool(
		MaxThread(BenchAntsSize),
		WithExpiryDuration(DefaultExpiredTime),
	)
	defer p.Close()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < RunTimes; j++ {
			_ = p.AddTask(demoFunc)
		}
	}
	b.StopTimer()
}

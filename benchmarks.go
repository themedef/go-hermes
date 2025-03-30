package hermes

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

func BenchmarkParallelGet() {
	debug.SetGCPercent(20)
	const total = 10000000
	maxConcurrent := 1

	db := NewStore(Config{ShardCount: maxConcurrent})
	ctx := context.Background()

	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < total; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			key := fmt.Sprintf("key-%d", i)

			_, _ = db.Get(ctx, key)

		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	rps := float64(total) / elapsed.Seconds()
	fmt.Printf("Executed %d Get operations in %.2f seconds\n", total, elapsed.Seconds())
	fmt.Printf("RPS: %.2f\n", rps)
}

func BenchmarkParallelSet() {
	const total = 1000000
	maxConcurrent := 1

	db := NewStore(Config{ShardCount: maxConcurrent})
	ctx := context.Background()

	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < total; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			key := fmt.Sprintf("key-%d", i)
			_ = db.Set(ctx, key, "value", 0)
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	rps := float64(total) / elapsed.Seconds()
	fmt.Printf("Executed %d Set operations in %.2f seconds\n", total, elapsed.Seconds())
	fmt.Printf("RPS: %.2f\n", rps)
}

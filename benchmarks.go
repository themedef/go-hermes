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
	const total = 1000000
	maxConcurrent := 1

	db := NewStore(Config{ShardCount: maxConcurrent})
	ctx := context.Background()
	var addWG sync.WaitGroup

	start := time.Now()

	for i := 0; i < total; i++ {
		key := fmt.Sprintf("key-%d", i)
		_, _ = db.Get(ctx, key)
	}

	addWG.Wait()
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

	start := time.Now()

	for i := 0; i < total; i++ {
		key := fmt.Sprintf("key-%d", i)
		_ = db.Set(ctx, key, "value", 0)
	}

	elapsed := time.Since(start)
	rps := float64(total) / elapsed.Seconds()
	fmt.Printf("Executed %d Set operations in %.2f seconds\n", total, elapsed.Seconds())
	fmt.Printf("RPS: %.2f\n", rps)
}

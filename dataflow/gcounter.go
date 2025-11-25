package dataflow

import (
    "context"
    "fmt"
    "sync/atomic"
    "time"
)

type GoroutineCounter struct {
    count int32
}

func (gc *GoroutineCounter) Add(delta int) {
    atomic.AddInt32(&gc.count, int32(delta))
}

func (gc *GoroutineCounter) Done() {
    atomic.AddInt32(&gc.count, -1)
}

func (gc *GoroutineCounter) Count() int {
    return int(atomic.LoadInt32(&gc.count))
}

func MonitorGoroutines(ctx context.Context, interval time.Duration, gc *GoroutineCounter) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            fmt.Printf("Active goroutines: %d\n", gc.Count())
        case <-ctx.Done():
            return
        }
    }
}

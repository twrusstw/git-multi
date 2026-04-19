package util

import (
	"runtime"
	"sync"
)

// DefaultWorkers caps per-repo concurrency so we don't spawn 100+ git
// processes at once on a directory with many repos.
var DefaultWorkers = runtime.NumCPU() * 2

// ParallelMap runs fn on each item concurrently, bounded by workers, and
// returns the results in the same order as items. workers <= 0 uses DefaultWorkers.
func ParallelMap[T any, R any](items []T, workers int, fn func(T) R) []R {
	if workers <= 0 {
		workers = DefaultWorkers
	}
	results := make([]R, len(items))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for i, it := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, it T) {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = fn(it)
		}(i, it)
	}
	wg.Wait()
	return results
}

// ParallelDo runs fn on each item concurrently, bounded by workers.
// Use when fn already handles its own output (e.g. via ui.LockedPrint).
func ParallelDo[T any](items []T, workers int, fn func(T)) {
	if workers <= 0 {
		workers = DefaultWorkers
	}
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for _, it := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(it T) {
			defer wg.Done()
			defer func() { <-sem }()
			fn(it)
		}(it)
	}
	wg.Wait()
}

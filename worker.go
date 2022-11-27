package jpipe

import (
	"sync"

	"github.com/junitechnology/jpipe/options"
)

type worker[T any, R any] func(node workerNode[T, R])

func (worker worker[T, R]) Pooled(opts ...options.PooledWorkerOption) worker[T, R] {
	concurrent := getOptionOrDefault(opts, Concurrent(1))
	if concurrent.Concurrency == 1 {
		return worker
	}

	return worker.pooled(concurrent.Concurrency)
}

func (worker worker[T, R]) pooled(concurrency int) worker[T, R] {
	return func(node workerNode[T, R]) {
		var wg sync.WaitGroup
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					node.HandlePanic()
				}()

				worker(node)
			}()
		}
		wg.Wait()
	}
}

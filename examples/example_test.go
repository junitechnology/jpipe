package examples

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
)

func TestWithGo(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	ids := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	channel := make(chan int)
	concurrency := 5
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				// The nested select gives priority to the ctx.Done() signal, so we always exit early if needed
				// Without it, a single select just has no priority, so a new value could be processed even if the context has been canceled
				case <-ctx.Done():
					return
				default:
					select {
					case id, open := <-channel:
						if !open {
							return
						}
						expensiveIOOperation(id)
					case <-ctx.Done(): // always check ctx.Done() to avoid
						return
					}
				}
			}
		}()
	}

	for _, id := range ids {
		channel <- id
	}
	close(channel)

	wg.Wait()

	end := time.Now()
	fmt.Printf("Total time: %s\n", end.Sub(start))
}

func TestWithJPipe(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	ids := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	pipeline := jpipe.NewPipeline(jpipe.Config{Context: ctx})
	<-jpipe.FromSlice(pipeline, ids).
		ForEach(expensiveIOOperation, jpipe.Concurrent(5))

	end := time.Now()
	fmt.Printf("Total time: %s\n", end.Sub(start))
}

func expensiveIOOperation(id int) {
	fmt.Printf("Doing expensive IO operation for id %d\n", id)
	time.Sleep(time.Second)
}

package examples

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gitlab.com/junitechnology/jpipe"
)

func TestWithGo(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	channel := make(chan int)
	concurrency := 10
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				select {
				case id := <-channel:
					{
						if id%2 == 1 {
							downloadInvoice(id)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	for _, i := range slice {
		channel <- i
	}

	wg.Wait()

	end := time.Now()
	fmt.Printf("Total time: %s\n", end.Sub(start))
}

func TestWithJPipe(t *testing.T) {
	ctx := context.Background()
	start := time.Now()

	pipeline := jpipe.NewPipeline(jpipe.Config{Context: ctx})
	jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
		Filter(func(i int) bool { return i%2 == 1 }).
		ForEach(downloadInvoice, jpipe.ForEachOptions{Concurrency: 10})

	end := time.Now()
	fmt.Printf("Total time: %s\n", end.Sub(start))
}

func downloadInvoice(id int) {
	fmt.Printf("Downloading invoice %d\n", id)
	time.Sleep(time.Second)
}

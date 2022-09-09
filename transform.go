package jpipe

import (
	"sync"
	"time"
)

type MapOptions struct {
	Concurrency int
	//Ordered bool
}

// Map transforms every input value with a mapper function and sends the results to the output channel.
//
// Example:
//  func(i int) int { return i + 10 }
//
// input : 0--1--2--3--4--5--X
//
// output: 10-11-12-13-14-15-X
func Map[T any, R any](input *Channel[T], mapper func(T) R, options ...MapOptions) *Channel[R] {
	opts := getOptions(MapOptions{Concurrency: 1}, options)
	worker := func(node workerNode[T, R]) {
		var wg sync.WaitGroup
		for i := 0; i < opts.Concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				node.LoopInput(0, func(value T) bool {
					return node.Send(mapper(value))
				})
			}()
		}
		wg.Wait()
	}

	_, output := newLinearPipelineNode("Map", input, 0, worker)
	return output
}

type FlatMapOptions struct {
	Concurrency int
	//Ordered bool
}

// Map transforms every input value into a Channel and for each of those, it sends all values to the output channel.
//
// Example:
//  func(i int) *Channel[int] { return FromSlice([]int{i, i + 10}) }
//
// input : 0------1------2------3------4------5------X
//
// output: 0-10---1-11---2-12---3-13---4-14---5-15---X
func FlatMap[T any, R any](input *Channel[T], mapper func(T) *Channel[R], options ...FlatMapOptions) *Channel[R] {
	opts := getOptions(FlatMapOptions{Concurrency: 1}, options)
	worker := func(node workerNode[T, R]) {
		var wg sync.WaitGroup
		for i := 0; i < opts.Concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				node.LoopInput(0, func(value T) bool {
					mappedChannel := mapper(value)
					loopOverChannel(node, mappedChannel.getChannel(), func(outputValue R) bool {
						if !node.Send(outputValue) {
							mappedChannel.unsubscribe()
							return false
						}
						return true
					})
					return true
				})
			}()
		}
		wg.Wait()
	}

	_, output := newLinearPipelineNode("FlatMap", input, 0, worker)
	return output
}

type BatchOptions struct {
	Size    int
	Timeout time.Duration
}

// Batch batches input values in slices and sends those slices to the output channel
// Batches can be limited by size with BatchOptions.Size and by time with BatchOptions.Timeout.
// It's possible to use size-only, time-only or size-and-time strategies.
//
// Example with BatchOptions{Size:3}:
//
// input : 0--1----2----------3------4--5----------6--7----X
//
// output: --------{1-2-3}--------------{3-4-5}-------{6-7}X
func Batch[T any](input *Channel[T], options ...BatchOptions) *Channel[[]T] {
	opts := getOptions(BatchOptions{}, options)
	nextTimeout := func() <-chan time.Time {
		if opts.Timeout > 0 {
			return time.After(opts.Timeout)
		}
		return make(<-chan time.Time)
	}

	worker := func(node workerNode[T, []T]) {
		batch := []T{}
		timeout := nextTimeout()
		for {
			var flush, done bool
			select {
			case <-node.QuitSignal(): // the nested select gives priority to the quit signal, so we always exit early if needed
				flush = true
				done = true
			default:
				select {
				case <-node.QuitSignal():
					flush = true
					done = true
				case value, open := <-node.Inputs()[0].getChannel():
					if !open {
						flush = true
						done = true
						break
					}
					batch = append(batch, value)
					if len(batch) == opts.Size {
						flush = true
					}
				case <-timeout:
					flush = true
				}
			}

			if flush {
				if !node.Send(batch) {
					return
				}
				batch = []T{}
				timeout = nextTimeout()
			}
			if done {
				return
			}
		}
	}

	_, output := newLinearPipelineNode("Batch", input, 0, worker)
	return output
}

package jpipe

import (
	"sync"
	"time"

	"github.com/junitechnology/jpipe/item"
	"github.com/junitechnology/jpipe/options"
)

// Map transforms every input value with a mapper function and sends the results to the output channel.
//
// Example:
//
//  output := Map(input, func(i int) int { return i + 10 })
//
//  input : 0--1--2--3--4--5--X
//  output: 10-11-12-13-14-15-X
func Map[T any, R any](input *Channel[T], mapper func(T) R, opts ...options.MapOptions) *Channel[R] {
	concurrent := getOptions(opts, Concurrent(1))
	worker := func(node workerNode[T, R]) {
		var wg sync.WaitGroup
		for i := 0; i < concurrent.Concurrency; i++ {
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

// FlatMap transforms every input value into a Channel and for each of those, it sends all values to the output channel.
//
// Example:
//
//  output := FlatMap(input, func(i int) *Channel[int] { return FromSlice([]int{i, i + 10}) })
//
//  input : 0------1------2------3------4------5------X
//  output: 0-10---1-11---2-12---3-13---4-14---5-15---X
func FlatMap[T any, R any](input *Channel[T], mapper func(T) *Channel[R], opts ...options.FlatMapOptions) *Channel[R] {
	concurrent := getOptions(opts, Concurrent(1))
	worker := func(node workerNode[T, R]) {
		var wg sync.WaitGroup
		for i := 0; i < concurrent.Concurrency; i++ {
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

// Batch batches input values in slices and sends those slices to the output channel
// Batches can be limited by size and by time.
// Size/time are ignored if they are 0
//
// Example:
//
//  output := Batch(input, 3, 0)
//
//  input : 0--1----2----------3------4--5----------6--7----X
//  output: --------{1-2-3}--------------{3-4-5}-------{6-7}X
func Batch[T any](input *Channel[T], size int, timeout time.Duration) *Channel[[]T] {
	nextTimeout := func() <-chan time.Time {
		if timeout > 0 {
			return time.After(timeout)
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
					if len(batch) == size {
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

// Wrap wraps every input value T in an Item[T] and sends it to the output channel.
// Item[T] is used mostly to represent items that can have either a value or an error.
// Another use for Item[T] is using the Context in it and enrich it in successive operators.
func Wrap[T any](input *Channel[T]) *Channel[item.Item[T]] {
	worker := func(node workerNode[T, item.Item[T]]) {
		node.LoopInput(0, func(value T) bool {
			return node.Send(item.Item[T]{Value: value})
		})
	}

	_, output := newLinearPipelineNode("Wrap", input, 0, worker)
	return output
}

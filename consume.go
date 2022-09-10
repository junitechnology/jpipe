package jpipe

import (
	"sync"
)

type ForEachOptions struct {
	Concurrency int
}

// ForEach calls the function passed as parameter for every value coming from the input channel.
// It returns a channel that will close when all input values have been processed.
func (channel *Channel[T]) ForEach(function func(T), options ...ForEachOptions) <-chan struct{} {
	opts := getOptions(ForEachOptions{Concurrency: 1}, options)
	worker := func(node workerNode[T, any]) {
		var wg sync.WaitGroup

		for i := 0; i < opts.Concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				node.LoopInput(0, func(value T) bool {
					function(value)
					return true
				})
			}()
		}
		wg.Wait()
	}

	node := newSinkPipelineNode("ForEach", channel, worker)
	return node.Done()
}

type ReduceOptions[R any] struct {
	InitialState R
}

// Reduce performs a stateful reduction of the input values.
// The reducer receives the current state and the current value, and must return the new state.
// The returned channel will return the final state when all input values have been processed.
//
// Example. Calculating the sum of all input values:
//
//   Reduce(input, func(acc int64, value int) int64 { return acc + int64(value) })
func Reduce[T any, R any](channel *Channel[T], reducer func(R, T) R, options ...ReduceOptions[R]) <-chan R {
	opts := getOptions(ReduceOptions[R]{}, options)
	resultChannel := make(chan R, 1)
	worker := func(node workerNode[T, any]) {
		var state R = opts.InitialState
		defer func() { resultChannel <- state }()
		node.LoopInput(0, func(value T) bool {
			state = reducer(state, value)
			return true
		})
	}

	newSinkPipelineNode("Reduce", channel, worker)
	return resultChannel
}

// ToSlice puts all values coming from the input channel in a slice.
// It returns a channel that will return the slice when all input values have been processed.
func (channel *Channel[T]) ToSlice() <-chan []T {
	resultChannel := make(chan []T, 1)
	worker := func(node workerNode[T, any]) {
		slice := []T{}
		defer func() { resultChannel <- slice }()
		node.LoopInput(0, func(value T) bool {
			slice = append(slice, value)
			return true
		})
	}

	newSinkPipelineNode("ToSlice", channel, worker)
	return resultChannel
}

type ToMapOptions struct {
	Keep KeepStrategy
}

// ToMap puts all values coming from the input channel in a map, using the getKey parameter to calculate the key.
// It returns a channel that will return the map when all input values have been processed.
func ToMap[T any, K comparable](channel *Channel[T], getKey func(T) K, options ...ToMapOptions) <-chan map[K]T {
	opts := getOptions(ToMapOptions{Keep: KEEP_FIRST}, options)
	resultChannel := make(chan map[K]T, 1)
	worker := func(node workerNode[T, any]) {
		resultMap := map[K]T{}
		defer func() { resultChannel <- resultMap }()
		node.LoopInput(0, func(value T) bool {
			key := getKey(value)
			if opts.Keep == KEEP_FIRST {
				if _, ok := resultMap[key]; ok {
					return true
				}
			}
			resultMap[key] = value
			return true
		})
	}

	newSinkPipelineNode("ToMap", channel, worker)
	return resultChannel
}

// ToGoChannel sends all values from the input channel to the returned Go channel.
// The returned Go channel closes when all input values have been processed.
func (channel *Channel[T]) ToGoChannel() <-chan T {
	goChannel := make(chan T)
	worker := func(node workerNode[T, any]) {
		defer close(goChannel)
		node.LoopInput(0, func(value T) bool {
			select {
			case <-node.QuitSignal():
				return false // the nested select gives priority to the quit signal, so we always exit early if needed
			default:
				select {
				case <-node.QuitSignal():
					return false
				case goChannel <- value:
				}
			}
			return true
		})
	}

	newSinkPipelineNode("ToGoChannel", channel, worker)
	return goChannel
}

// Last returns a channel that will return the last input value received, immediately after the input channel has been closed.
func (channel *Channel[T]) Last() <-chan T {
	resultChannel := make(chan T, 1)
	worker := func(node workerNode[T, any]) {
		var lastValue T
		var ok bool
		defer func() {
			if ok {
				resultChannel <- lastValue
			} else {
				close(resultChannel)
			}
		}()
		node.LoopInput(0, func(value T) bool {
			lastValue = value
			ok = true
			return true
		})
	}

	newSinkPipelineNode("Last", channel, worker)
	return resultChannel
}

// Count counts input elements.
// The final count is sent to the return channel when all input values are red, or the pipeline is canceled.
//
func (channel *Channel[T]) Count() <-chan int64 {
	resultChannel := make(chan int64, 1)
	worker := func(node workerNode[T, any]) {
		count := int64(0)
		defer func() { resultChannel <- count }()
		node.LoopInput(0, func(value T) bool {
			count++
			return true
		})
	}

	newSinkPipelineNode("Count", channel, worker)
	return resultChannel
}

// Any returns a channel that will return true as soon as an input value satisfies the predicate.
// The channel will return false if all input values were processed and no one satisfied the predicate.
func (channel *Channel[T]) Any(predicate func(T) bool) <-chan bool {
	resultChannel := make(chan bool)
	worker := func(node workerNode[T, any]) {
		result := false
		defer func() { resultChannel <- result }()
		node.LoopInput(0, func(value T) bool {
			result = predicate(value)
			return !result
		})
	}

	newSinkPipelineNode("Any", channel, worker)
	return resultChannel
}

// All returns a channel that will return false as soon as an input value does not satisfy the predicate.
// The channel will return true if all input values were processed and they all satisfied the predicate.
func (channel *Channel[T]) All(predicate func(T) bool) <-chan bool {
	resultChannel := make(chan bool)
	worker := func(node workerNode[T, any]) {
		result := true
		defer func() { resultChannel <- result }()
		node.LoopInput(0, func(value T) bool {
			result = predicate(value)
			return result
		})
	}

	newSinkPipelineNode("All", channel, worker)
	return resultChannel
}

// None returns a channel that will return false as soon as an input value satisfies the predicate.
// The channel will return true if all input values were processed and no one satisfied the predicate.
func (channel *Channel[T]) None(predicate func(T) bool) <-chan bool {
	resultChannel := make(chan bool)
	worker := func(node workerNode[T, any]) {
		result := true
		defer func() { resultChannel <- result }()
		node.LoopInput(0, func(value T) bool {
			result = !predicate(value)
			return result
		})
	}

	newSinkPipelineNode("None", channel, worker)
	return resultChannel
}

package jpipe

import (
	"sync"

	"github.com/junitechnology/jpipe/options"
)

// ForEach calls the function passed as parameter for every value coming from the input channel.
// The returned channel will close when all input values have been processed, or the pipeline is canceled.
func (input *Channel[T]) ForEach(function func(T), opts ...options.ForEachOptions) <-chan struct{} {
	concurrent := getOptions(opts, Concurrent(1))
	worker := func(node workerNode[T, any]) {
		var wg sync.WaitGroup

		for i := 0; i < concurrent.Concurrency; i++ {
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

	node := newSinkPipelineNode("ForEach", input, worker)
	return node.Done()
}

// Reduce performs a stateful reduction of the input values.
// The reducer receives the current state and the current value, and must return the new state.
// The final state is sent to the returned channel when all input values have been processed, or the pipeline is canceled.
//
// Example. Calculating the sum of all input values:
//
//  output := Reduce(input, func(acc int64, value int) int64 { return acc + int64(value) })
//
//  input : 0--1--2--3--X
//  output: ------------6
func Reduce[T any, R any](input *Channel[T], reducer func(R, T) R) <-chan R {
	resultChannel := make(chan R, 1)
	worker := func(node workerNode[T, any]) {
		var state R
		defer func() { resultChannel <- state }()
		node.LoopInput(0, func(value T) bool {
			state = reducer(state, value)
			return true
		})
	}

	newSinkPipelineNode("Reduce", input, worker)
	return resultChannel
}

// ToSlice puts all values coming from the input channel in a slice.
// The resulting slice is sent to the returned channel when all input values have been processed, or the pipeline is canceled.
// The slice may have partial results if the pipeline failed, so you must remember to check the pipeline's Error() method.
//
// Example:
//
//  output := input.ToSlice()
//
//  input : 0--1--2--3--X
//  output: ------------{0,1,2,3}
func (input *Channel[T]) ToSlice() <-chan []T {
	resultChannel := make(chan []T, 1)
	worker := func(node workerNode[T, any]) {
		slice := []T{}
		defer func() { resultChannel <- slice }()
		node.LoopInput(0, func(value T) bool {
			slice = append(slice, value)
			return true
		})
	}

	newSinkPipelineNode("ToSlice", input, worker)
	return resultChannel
}

// ToMap puts all values coming from the input channel in a map, using the getKey parameter to calculate the key.
// The resulting map is sent to the returned channel when all input values have been processed, or the pipeline is canceled.
//
// Example:
//
//  output := ToMap(input, func(value string) string { return strings.Split(value, "_")[0] })
//
//  input : A_0--B_1--C_2--X
//  output: ---------------{A:A_0, B:B_1, C:C_2}
func ToMap[T any, K comparable](input *Channel[T], getKey func(T) K, opts ...options.ToMapOptions) <-chan map[K]T {
	keep := getOptions(opts, KeepFirst())
	resultChannel := make(chan map[K]T, 1)
	worker := func(node workerNode[T, any]) {
		resultMap := map[K]T{}
		defer func() { resultChannel <- resultMap }()
		node.LoopInput(0, func(value T) bool {
			key := getKey(value)
			if keep.Strategy == options.KEEP_FIRST {
				if _, ok := resultMap[key]; ok {
					return true
				}
			}
			resultMap[key] = value
			return true
		})
	}

	newSinkPipelineNode("ToMap", input, worker)
	return resultChannel
}

// ToGoChannel sends all values from the input channel to the returned Go channel.
// The returned Go channel closes when all input values have been processed, or the pipeline is canceled.
//
// Example:
//
//  output := input.ToGoChannel()
//
//  input : 0--1--2--3--X
//  output: 0--1--2--3--X
func (input *Channel[T]) ToGoChannel() <-chan T {
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

	newSinkPipelineNode("ToGoChannel", input, worker)
	return goChannel
}

// Last sends the last value received from the input channel to the output channel.
// The last value is sent to the returned channel when all input values have been processed, or the pipeline is canceled.
//
// Example:
//
//  output := input.Last()
//
//  input : 0--1--2--3------X
//  output: ----------------3
func (input *Channel[T]) Last() <-chan T {
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

	newSinkPipelineNode("Last", input, worker)
	return resultChannel
}

// Count counts input values and sends the final count to the output channel.
// The final count is sent to the return channel when all input values have been processed, or the pipeline is canceled.
//
// Example:
//
//  output := input.ToGoChannel()
//
//  input : 9--8--7--6--X
//  output: ------------4
func (input *Channel[T]) Count() <-chan int64 {
	resultChannel := make(chan int64, 1)
	worker := func(node workerNode[T, any]) {
		count := int64(0)
		defer func() { resultChannel <- count }()
		node.LoopInput(0, func(value T) bool {
			count++
			return true
		})
	}

	newSinkPipelineNode("Count", input, worker)
	return resultChannel
}

// Any determines if any input value matches the predicate.
// If no value matches the predicate, false is sent to the returned channel when all input values have been processed, or the pipeline is canceled.
// If instead some value is found to match the predicate, true is immediately sent to the returned channel and no more input values are read.
//
// Example 1:
//
//  output := input.Any(func(value int) bool { return value > 3 })
//
//  input : 0--1--2--3--X
//  output: ------------false
//
// Example 2:
//
//  output := input.Any(func(value int) bool { return value >= 2 })
//
//  input : 0--1--2--3--X
//  output: ------true
func (input *Channel[T]) Any(predicate func(T) bool) <-chan bool {
	resultChannel := make(chan bool)
	worker := func(node workerNode[T, any]) {
		result := false
		defer func() { resultChannel <- result }()
		node.LoopInput(0, func(value T) bool {
			result = predicate(value)
			return !result
		})
	}

	newSinkPipelineNode("Any", input, worker)
	return resultChannel
}

// All determines if all input values match the predicate.
// If all values match the predicate, true is sent to the returned channel when all input values have been processed, or the pipeline is canceled.
// If instead some value does not match the predicate, false is immediately sent to the returned channel and no more input values are read.
//
// Example 1:
//
//  output := input.All(func(value int) bool { return value < 4 })
//
//  input : 0--1--2--3--X
//  output: ------------true
//
// Example 2:
//
//  output := input.All(func(value int) bool { return value < 2 })
//
//  input : 0--1--2--3--X
//  output: ------false
func (input *Channel[T]) All(predicate func(T) bool) <-chan bool {
	resultChannel := make(chan bool)
	worker := func(node workerNode[T, any]) {
		result := true
		defer func() { resultChannel <- result }()
		node.LoopInput(0, func(value T) bool {
			result = predicate(value)
			return result
		})
	}

	newSinkPipelineNode("All", input, worker)
	return resultChannel
}

// None determines if no input value matches the predicate.
// If no value matches the predicate, true is sent to the returned channel when all input values have been processed, or the pipeline is canceled.
// If instead some value matches the predicate, false is immediately sent to the returned channel and no more input values are read.
//
// Example 1:
//
//  output := input.None(func(value int) bool { return value > 3 })
//
//  input : 0--1--2--3--X
//  output: ------------true
//
// Example 2:
//
//  output := input.None(func(value int) bool { return value >= 2 })
//
//  input : 0--1--2--3--X
//  output: ------false
func (input *Channel[T]) None(predicate func(T) bool) <-chan bool {
	resultChannel := make(chan bool)
	worker := func(node workerNode[T, any]) {
		result := true
		defer func() { resultChannel <- result }()
		node.LoopInput(0, func(value T) bool {
			result = !predicate(value)
			return result
		})
	}

	newSinkPipelineNode("None", input, worker)
	return resultChannel
}

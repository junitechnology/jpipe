package jpipe

import (
	"github.com/junitechnology/jpipe/options"
)

// ForEach calls the function passed as parameter for every value coming from the input channel.
// The returned channel will close when all input values have been processed, or the pipeline is canceled.
func (input *Channel[T]) ForEach(function func(T), opts ...options.ForEachOption) <-chan struct{} {
	var worker worker[T, any] = func(node workerNode[T, any]) {
		node.LoopInput(0, func(value T) bool {
			function(value)
			return true
		})
	}
	worker = worker.Pooled(getPooledWorkerOptions(opts)...)

	node := newSinkPipelineNode("ForEach", input, worker, getNodeOptions(opts)...)
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
	var state R
	worker := func(node workerNode[T, any]) {
		node.LoopInput(0, func(value T) bool {
			state = reducer(state, value)
			return true
		})
	}

	node := newSinkPipelineNode("Reduce", input, worker)
	return resultChannel(node, func(ch chan R) { ch <- state })
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
	slice := []T{}
	worker := func(node workerNode[T, any]) {
		node.LoopInput(0, func(value T) bool {
			slice = append(slice, value)
			return true
		})
	}

	node := newSinkPipelineNode("ToSlice", input, worker)
	return resultChannel(node, func(ch chan []T) { ch <- slice })
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
func ToMap[T any, K comparable](input *Channel[T], getKey func(T) K, opts ...options.ToMapOption) <-chan map[K]T {
	keep := getOptionOrDefault(opts, KeepFirst())
	resultMap := map[K]T{}
	worker := func(node workerNode[T, any]) {
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

	node := newSinkPipelineNode("ToMap", input, worker, getNodeOptions(opts)...)
	return resultChannel(node, func(ch chan map[K]T) { ch <- resultMap })
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

	node := newSinkPipelineNode("ToGoChannel", input, worker)
	go func() {
		<-node.Done()
		close(goChannel)
	}()
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
	var gotValue bool
	var lastValue T
	worker := func(node workerNode[T, any]) {
		node.LoopInput(0, func(value T) bool {
			gotValue = true
			lastValue = value
			return true
		})
	}

	node := newSinkPipelineNode("Last", input, worker)
	return resultChannel(node, func(ch chan T) {
		if gotValue {
			ch <- lastValue
		} else {
			close(ch)
		}
	})
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
	count := int64(0)
	worker := func(node workerNode[T, any]) {
		node.LoopInput(0, func(value T) bool {
			count++
			return true
		})
	}

	node := newSinkPipelineNode("Count", input, worker)
	return resultChannel(node, func(ch chan int64) { ch <- count })
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
	result := false
	worker := func(node workerNode[T, any]) {
		node.LoopInput(0, func(value T) bool {
			result = predicate(value)
			return !result
		})
	}

	node := newSinkPipelineNode("Any", input, worker)
	return resultChannel(node, func(ch chan bool) { ch <- result })
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
	result := true
	worker := func(node workerNode[T, any]) {
		node.LoopInput(0, func(value T) bool {
			result = predicate(value)
			return result
		})
	}

	node := newSinkPipelineNode("All", input, worker)
	return resultChannel(node, func(ch chan bool) { ch <- result })
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
	result := true
	worker := func(node workerNode[T, any]) {
		node.LoopInput(0, func(value T) bool {
			result = !predicate(value)
			return result
		})
	}

	node := newSinkPipelineNode("None", input, worker)
	return resultChannel(node, func(ch chan bool) { ch <- result })
}

func resultChannel[R any](node pipelineNode, doOnChannel func(chan R)) <-chan R {
	resultCh := make(chan R, 1)
	go func() {
		<-node.Done()
		doOnChannel(resultCh)
	}()

	return resultCh
}

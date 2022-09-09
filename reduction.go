package jpipe

import "golang.org/x/exp/constraints"

type ReduceOptions[S any] struct {
	InitialState S
}

// Reduce performs a stateful reduction of the input values.
//
// Notice that the reducer function is richer than in most languages/libraries.
// It takes makes a distinction between the state S, which is the accumulated value and the return value R.
// S and R typically are the same and that's why they are conflated in a single value in most languages/libraries, but having them separate allows for more complex use cases.
//
// Also, unlike most reduce implementations out there, this one doesn't just return a single final value.
// It returns a channel with the series of all reduced values.
// This allows for implementing things like a moving average.
//
// Example:
//  Reduce(channel, func(state int, value int) (int, int) { return state + value, state + value })
//
// input : 0--1--2--3--4--5--X
//
// output: 0--1--3--6--10-15-X
func Reduce[T any, S any, R any](input *Channel[T], reducer func(S, T) (S, R), options ...ReduceOptions[S]) *Channel[R] {
	opts := getOptions(ReduceOptions[S]{}, options)
	worker := func(node workerNode[T, R]) {
		var state S = opts.InitialState
		node.LoopInput(0, func(value T) bool {
			var result R
			state, result = reducer(state, value)
			return node.Send(result)
		})
	}

	_, output := newLinearPipelineNode("Reduce", input, 0, worker)
	return output
}

type MaxByOptions struct {
	Keep KeepStrategy
}

// MaxBy reads every input value, and for each value it sends the maximum value seen up to that point.
// The ordering between elements is established by an ordering key, calculated for each value with toKey.
//
// Example:
//  func(value T) O { return value.x }
//
// input : {x:0}--{x:2}--{x:1}--{x:5}--{x:3}--{x:9}--X
//
// output: {x:0}--{x:2}--{x:2}--{x:5}--{x:5}--{x:9}--X
func MaxBy[T any, O constraints.Ordered](input *Channel[T], toKey func(T) O, options ...MaxByOptions) *Channel[T] {
	opts := getOptions(MaxByOptions{Keep: KEEP_FIRST}, options)
	worker := func(node workerNode[T, T]) {
		var maxKey O
		var maxValue *T
		node.LoopInput(0, func(value T) bool {
			key := toKey(value)
			if maxValue == nil || key > maxKey || (opts.Keep == KEEP_LAST && key == maxKey) {
				maxKey = key
				maxValue = &value
			}
			return node.Send(*maxValue)
		})
	}

	_, output := newLinearPipelineNode("MaxBy", input, 0, worker)
	return output
}

type MinByOptions struct {
	Keep KeepStrategy
}

// MinBy reads every input value, and for each value it sends the minimum value seen up to that point.
// The ordering between values is established by an ordering key, calculated for each value with toKey.
//
// Example:
//   func(value T) O { return value.x }:
//
// input : {x:9}--{x:6}--{x:7}--{x:3}--{x:5}--{x:1}--X
//
// output: {x:9}--{x:6}--{x:6}--{x:3}--{x:3}--{x:1}---X
func MinBy[T any, O constraints.Ordered](input *Channel[T], toKey func(T) O, options ...MinByOptions) *Channel[T] {
	opts := getOptions(MinByOptions{Keep: KEEP_FIRST}, options)
	worker := func(node workerNode[T, T]) {
		var maxKey O
		var maxValue *T
		node.LoopInput(0, func(value T) bool {
			key := toKey(value)
			if maxValue == nil || key < maxKey || (opts.Keep == KEEP_LAST && key == maxKey) {
				maxKey = key
				maxValue = &value
			}
			return node.Send(*maxValue)
		})
	}

	_, output := newLinearPipelineNode("MinBy", input, 0, worker)
	return output
}

// Count counts input elements and publishes the counts to the output channel
//
// Example:
//
// input : a--b--a--b--a--b--X
//
// output: 1--2--3--4--5--6--X
func (input *Channel[T]) Count() *Channel[int64] {
	worker := func(node workerNode[T, int64]) {
		count := int64(0)
		node.LoopInput(0, func(value T) bool {
			count++
			return node.Send(count)
		})
	}

	_, output := newLinearPipelineNode("Count", input, 0, worker)
	return output
}

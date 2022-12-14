package jpipe

import (
	"time"

	"github.com/junitechnology/jpipe/item"
	"github.com/junitechnology/jpipe/options"
)

// Buffer transparently passes input values to the output channel, but the output channel is buffered.
// It is useful to avoid backpressure from slow consumers.
func (input *Channel[T]) Buffer(n int) *Channel[T] {
	worker := func(node workerNode[T, T]) {
		node.LoopInput(0, func(value T) bool {
			return node.Send(value)
		})
	}

	_, output := newLinearPipelineNode("Buffer", input, n, worker)
	return output
}

// Tap runs a function as a side effect for each input value, and then sends the input values transparently to the output channel.
// A common use case is logging.
func (input *Channel[T]) Tap(function func(T)) *Channel[T] {
	worker := func(node workerNode[T, T]) {
		node.LoopInput(0, func(value T) bool {
			function(value)
			return node.Send(value)
		})
	}

	_, output := newLinearPipelineNode("Tap", input, 0, worker)
	return output
}

// Interval transparently passes all input values to the output channel, but a time interval is awaited after each element before sending another one.
// No value is sent to the output while that interval is active.
// This operator is prone to generating backpressure, so use it with care, and consider adding a Buffer before it.
//
// Example(assume each hyphen is 1 ms):
//
//  output := input.Interval(4*time.Millisecond)
//
//  input : 0--1--2--------------3--4--5--X
//  output: 0----1----2----------3----4----5-X
func (input *Channel[T]) Interval(interval func(value T) time.Duration) *Channel[T] {
	worker := func(node workerNode[T, T]) {
		timeout := time.After(0)
		node.LoopInput(0, func(value T) bool {
			select {
			case <-node.QuitSignal():
				return false
			case <-timeout:
			}

			if !node.Send(value) {
				return false
			}
			timeout = time.After(interval(value))
			return true
		})
	}

	_, output := newLinearPipelineNode("Interval", input, 0, worker)
	return output
}

// Broadcast broadcasts each input value to every output channel.
// The next input value is not read by this operator until all output channels have read the current one.
// Bear in mind that if one of the output channels is a slow consumer, it may block the other consumers.
// This is a particularly annoying type of backpressure, cause not only does it block the input, it also blocks other consumers.
// To avoid this, consider using BroadcastOptions.BufferSize and the output channels will be buffered, with no need for an extra Buffer operator.
//
// Example (assume each hyphen is 1 ms):
//
//  outputs := input.Broadcast(2, Buffered(4*time.Millisecond))
//
//  input  : 0--1--2--3--4--5---X
//  output1: 0--1--2--3--4--5---X
//  output2: -0--1--2--3--4--5--X
func (input *Channel[T]) Broadcast(numOutputs int, opts ...options.BroadcastOptions) []*Channel[T] {
	buffered := getOptions(opts, Buffered(0))
	worker := func(node workerNode[T, T]) {
		node.LoopInput(0, func(value T) bool {
			return node.Send(value)
		})
	}

	_, outputs := newPipelineNode("Broadcast", input.getPipeline(), []*Channel[T]{input}, numOutputs, buffered.Size, worker)
	return outputs
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

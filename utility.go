package jpipe

import (
	"time"

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

	_, output := newLinearPipelineNode("Buffer", input, worker, Buffered(n))
	return output
}

// Tap runs a function as a side effect for each input value, and then sends the input values transparently to the output channel.
// A common use case is logging.
func (input *Channel[T]) Tap(function func(T), opts ...options.TapOption) *Channel[T] {
	var processor processor[T, T] = func(value T) (T, bool) {
		function(value)
		return value, true
	}
	worker := processor.PooledWorker(getPooledWorkerOptions(opts)...)

	_, output := newLinearPipelineNode("Tap", input, worker)
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

	_, output := newLinearPipelineNode("Interval", input, worker)
	return output
}

package jpipe

import (
	"golang.org/x/exp/constraints"
)

// FromGoChannel creates a Channel from a Go channel
func FromGoChannel[T any](pipeline *Pipeline, channel <-chan T) *Channel[T] {
	worker := func(node workerNode[any, T]) {
		loopOverChannel(node, channel, func(value T) bool {
			return node.Send(value)
		})
	}

	_, output := newSourcePipelineNode("FromGoChannel", pipeline, 0, worker)
	return output
}

// FromSlice creates a Channel from a slice.
// All values in the slice are sent to the channel in order
func FromSlice[T any](pipeline *Pipeline, slice []T) *Channel[T] {
	worker := func(node workerNode[any, T]) {
		for _, value := range slice {
			if !node.Send(value) {
				return
			}
		}
	}

	_, output := newSourcePipelineNode("FromSlice", pipeline, 0, worker)
	return output
}

// FromRange creates a Channel from a range of integers.
// All integers between start and end (both inclusive) are sent to the channel in order
func FromRange[T constraints.Integer](pipeline *Pipeline, start T, end T) *Channel[T] {
	worker := func(node workerNode[any, T]) {
		for i := start; i <= end; i++ {
			if !node.Send(i) {
				return
			}
		}
	}

	_, output := newSourcePipelineNode("FromRange", pipeline, 0, worker)
	return output
}

// FromGenerator creates a Channel from a stateless generator function.
// Values returned by the function are sent to the channel in order.
func FromGenerator[T any](pipeline *Pipeline, generator func(i uint64) T) *Channel[T] {
	worker := func(node workerNode[any, T]) {
		for i := uint64(0); ; i++ {
			if !node.Send(generator(i)) {
				return
			}
		}
	}

	_, output := newSourcePipelineNode("FromGenerator", pipeline, 0, worker)
	return output
}

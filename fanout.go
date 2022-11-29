package jpipe

import "github.com/junitechnology/jpipe/options"

// Split sends each input value to any of the output channels, with no specific priority.
//
// Example:
//
//  outputs := input.Split(2, Buffered(4))
//
//  input  : 0--1--2--3--4--5---X
//  output1: 0-----2--3-----5---X
//  output2: ---1--------4------X
func (input *Channel[T]) Split(numOutputs int, opts ...options.SplitOption) []*Channel[T] {
	worker := func(node workerNode[T, T]) {
		node.LoopInput(0, func(value T) bool {
			return node.Send(value)
		})
	}

	_, outputs := newPipelineNode("Split", input.getPipeline(), []*Channel[T]{input}, numOutputs, worker, true, getNodeOptions(opts)...)
	return outputs
}

// Broadcast sends each input value to every output channel.
// The next input value is not read by this operator until all output channels have read the current one.
// Bear in mind that if one of the output channels is a slow consumer, it may block the other consumers.
// This is a particularly annoying type of backpressure, cause not only does it block the input, it also blocks other consumers.
// To avoid this, consider using options.Buffered and the output channels will be buffered, with no need for an extra Buffer operator.
//
// Example:
//
//  outputs := input.Broadcast(2, Buffered(4))
//
//  input  : 0--1--2--3--4--5---X
//  output1: 0--1--2--3--4--5---X
//  output2: 0--1--2--3--4--5---X
func (input *Channel[T]) Broadcast(numOutputs int, opts ...options.BroadcastOption) []*Channel[T] {
	worker := func(node workerNode[T, T]) {
		node.LoopInput(0, func(value T) bool {
			return node.Send(value)
		})
	}

	_, outputs := newPipelineNode("Broadcast", input.getPipeline(), []*Channel[T]{input}, numOutputs, worker, false, getNodeOptions(opts)...)
	return outputs
}

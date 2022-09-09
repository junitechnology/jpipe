package jpipe

import (
	"sync"
)

// Merge merges multiple input channels to a single output channel. Values from input
// channels are sent to the output channels as they arrive, with no specific priority.
//
// Example:
//
// input1: 0----1----2------3-X
//
// input2: -----5------6------X
//
// output: 0----5-1--2-6----3-X
func Merge[T any](inputs ...*Channel[T]) *Channel[T] {
	worker := func(node workerNode[T, T]) {
		var wg sync.WaitGroup
		for i := range inputs {
			i := i // avoid goroutine capturing the loop i
			wg.Add(1)
			go func() {
				defer wg.Done()
				node.LoopInput(i, func(value T) bool {
					return node.Send(value)
				})
			}()
		}
		wg.Wait()
	}

	_, output := newPipelineNode("Merge", inputs[0].getPipeline(), inputs, 1, 0, worker)
	return output[0]
}

// Concat concatenates multiple input channels to a single output channel.
// Channels are consumed in order, e.g., the second channel won't be consumed
// until the first channel is closed.
//
// Example:
//
// input 1: 0----1----2------3-X
//
// input 2: -----5------6--------------7--X
//
// output : 0----1----2------3-5-6-----7--X
func Concat[T any](inputs ...*Channel[T]) *Channel[T] {
	worker := func(node workerNode[T, T]) {
		for i := range inputs {
			node.LoopInput(i, func(value T) bool {
				return node.Send(value)
			})
		}
	}

	_, output := newPipelineNode("Concat", inputs[0].getPipeline(), inputs, 1, 0, worker)
	return output[0]
}

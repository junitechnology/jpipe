package jpipe

import (
	"sync"

	"github.com/junitechnology/jpipe/options"
)

type processor[T any, R any] func(value T) (outputValue R, send bool)

func (processor processor[T, R]) PooledWorker(opts ...options.PooledWorkerOption) worker[T, R] {
	concurrent := getOptionOrDefault(opts, Concurrent(1))
	ordered := getOption[options.PooledWorkerOption, options.Ordered](opts)

	if concurrent.Concurrency == 1 {
		return processor.singleLoopWorker()
	} else if ordered == nil {
		return processor.singleLoopWorker().Pooled(concurrent)
	}

	return processor.orderedPooledWorker(concurrent.Concurrency, ordered.OrderBufferSize)
}

func (processor processor[T, R]) singleLoopWorker() worker[T, R] {
	return func(node workerNode[T, R]) {
		node.LoopInput(0, func(value T) bool {
			if output, send := processor(value); send {
				return node.Send(output)
			}
			return true
		})
	}
}

func (processor processor[T, R]) orderedPooledWorker(concurrency int, orderBufferSize int) worker[T, R] {
	return func(node workerNode[T, R]) {
		internalInput := make(chan orderedValue[T])
		orderingBuffer := NewOrderingBuffer(node, concurrency, orderBufferSize)

		var wg sync.WaitGroup
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					node.HandlePanic()
				}()

				loopOverChannel(node, internalInput, func(value orderedValue[T]) bool {
					output, send := processor(*value.value)
					return orderingBuffer.Send(orderedValue[R]{value: &output, idx: value.idx, skip: !send})
				})
			}()
		}
		orderingBuffer.Start()

		idx := int64(0)
		node.LoopInput(0, func(value T) bool {
			internalInput <- orderedValue[T]{value: &value, idx: idx}
			idx++
			return true
		})
		close(internalInput)

		wg.Wait()

		orderingBuffer.Stop()
		<-orderingBuffer.Done()
	}
}

type orderedValue[T any] struct {
	value *T
	idx   int64
	skip  bool
}

type orderingBuffer[T any, R any] struct {
	node     workerNode[T, R]
	input    chan orderedValue[R]
	feedback chan struct{}
	done     chan struct{}
}

func NewOrderingBuffer[T any, R any](node workerNode[T, R], concurrency int, bufferSize int) *orderingBuffer[T, R] {
	return &orderingBuffer[T, R]{
		node:     node,
		input:    make(chan orderedValue[R], bufferSize+concurrency),
		feedback: make(chan struct{}, bufferSize),
		done:     make(chan struct{}),
	}
}

func (b *orderingBuffer[T, R]) Send(value orderedValue[R]) bool {
	b.input <- value

	select {
	case <-b.feedback:
		return true
	case <-b.node.QuitSignal():
		return false
	}
}

func (b *orderingBuffer[T, R]) Start() {
	go func() {
		for i := 0; i < cap(b.feedback); i++ {
			b.feedback <- struct{}{}
		}

		buffer := map[int64]orderedValue[R]{}
		currentIndex := int64(0)
		loopOverChannel(b.node, b.input, func(value orderedValue[R]) bool {
			buffer[value.idx] = value
			for {
				value, ok := buffer[currentIndex]
				if !ok {
					return true
				}
				delete(buffer, currentIndex)

				if !value.skip {
					sent := b.node.Send(*value.value)
					if !sent {
						return false
					}
				}
				currentIndex++
				b.feedback <- struct{}{}
			}
		})

		close(b.done)
	}()
}

func (b *orderingBuffer[T, R]) Stop() {
	close(b.input)
}

func (b *orderingBuffer[T, R]) Done() <-chan struct{} {
	return b.done
}

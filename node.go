package jpipe

import (
	"fmt"
	"sync"
)

type pipelineNode interface {
	Start()
	Done() <-chan struct{}
	Children() []pipelineNode
	IsSource() bool
	IsSink() bool
}

type node[T any, R any] struct {
	lock sync.Mutex

	nodeType string

	pipeline        *Pipeline
	inputs          []*Channel[T]
	outputs         []*Channel[R]
	outputWriters   []chan<- R
	subscriptions   []chan struct{}
	allUnsubscribed chan struct{}

	worker      func(workerNode[T, R])
	concurrency int
	quitSignal  chan struct{}
	doneSignal  chan struct{}
}

type workerNode[T any, R any] interface {
	Inputs() []*Channel[T]
	LoopInput(i int, function func(value T) bool)
	Send(value R) bool
	QuitSignal() <-chan struct{}
}

func newPipelineNode[T any, R any](
	nodeType string,
	pipeline *Pipeline,
	inputs []*Channel[T],
	numOutputs int,
	outputChannelSize int,
	worker func(workerNode[T, R]),
	concurrency int) (pipelineNode, []*Channel[R]) {

	node := &node[T, R]{
		nodeType:        nodeType,
		pipeline:        pipeline,
		inputs:          inputs,
		outputs:         make([]*Channel[R], numOutputs),
		outputWriters:   make([]chan<- R, numOutputs),
		subscriptions:   make([]chan struct{}, numOutputs),
		allUnsubscribed: make(chan struct{}),
		worker:          worker,
		quitSignal:      make(chan struct{}),
		doneSignal:      make(chan struct{}),
		concurrency:     concurrency,
	}

	for _, input := range inputs {
		input.setToNode(node)
	}

	for i := 0; i < numOutputs; i++ {
		goChannel := make(chan R, outputChannelSize)
		n := i // avoid capturing the loop var i in the unsubscriber closure
		node.outputs[i] = newChannel(pipeline, goChannel, func() { node.unsubscribe(n) })
		node.outputWriters[i] = goChannel
		node.subscriptions[i] = make(chan struct{})
	}

	pipeline.addNode(node)

	return node, node.outputs
}

func newLinearPipelineNode[T any, R any](nodeType string, input *Channel[T], outputChannelSize int, worker func(workerNode[T, R]), concurrency int) (pipelineNode, *Channel[R]) {
	node, outputs := newPipelineNode(nodeType, input.getPipeline(), []*Channel[T]{input}, 1, outputChannelSize, worker, concurrency)
	return node, outputs[0]
}

func newSourcePipelineNode[R any](nodeType string, pipeline *Pipeline, outputChannelSize int, worker func(workerNode[any, R]), concurrency int) (pipelineNode, *Channel[R]) {
	node, outputs := newPipelineNode(nodeType, pipeline, []*Channel[any]{}, 1, outputChannelSize, worker, concurrency)
	return node, outputs[0]
}

func newSinkPipelineNode[T any](nodeType string, input *Channel[T], worker func(workerNode[T, any]), concurrency int) pipelineNode {
	node, _ := newPipelineNode(nodeType, input.getPipeline(), []*Channel[T]{input}, 0, 0, worker, concurrency)
	return node
}

func (node *node[T, R]) Start() {
	go func() {
		defer func() {
			for i := range node.outputs {
				close(node.outputWriters[i])
			}
			for i := range node.inputs {
				node.inputs[i].unsubscribe()
			}
			close(node.doneSignal)
		}()

		go func() {
			select {
			case <-node.pipeline.Done():
			case <-node.allUnsubscribed:
			}
			close(node.quitSignal)
		}()

		var wg sync.WaitGroup
		for i := 0; i < node.concurrency; i++ {
			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
					if r := recover(); r != nil {
						node.pipeline.Cancel(fmt.Errorf("%v", r))
					}
				}()
				node.worker(node)
			}()
		}
		wg.Wait()
	}()
}

func (node *node[T, R]) QuitSignal() <-chan struct{} {
	return node.quitSignal
}

func (node *node[T, R]) Done() <-chan struct{} {
	return node.doneSignal
}

func (node *node[T, R]) Children() []pipelineNode {
	children := make([]pipelineNode, len(node.outputs))
	for i := range node.outputs {
		children[i] = node.outputs[i].getToNode()
	}

	return children
}

func (node *node[T, R]) IsSource() bool {
	return len(node.inputs) == 0
}

func (node *node[T, R]) IsSink() bool {
	return len(node.outputs) == 0
}

func (node *node[T, R]) Inputs() []*Channel[T] {
	return node.inputs
}

func (node *node[T, R]) LoopInput(i int, function func(value T) bool) {
	loopOverChannel[T, T, R](node, node.inputs[i].getChannel(), function)
}

func (node *node[T, R]) Send(value R) bool {
	success := false
	for i := range node.outputs {
		select {
		case <-node.quitSignal:
			return false // the nested select gives priority to the quit signal, so we always exit early if needed
		default:
			select {
			case <-node.quitSignal:
				return false
			case node.outputWriters[i] <- value:
				success = true // we return true if we sent to at least one subscriber. If we don't it means there's no active subscriber left.
			case <-node.subscriptions[i]: // do nothing if subscription is canceled
			}
		}
	}

	return success
}

func (node *node[T, R]) unsubscribe(n int) {
	node.lock.Lock()
	defer node.lock.Unlock()

	select {
	case <-node.subscriptions[n]:
		return // for idempotency
	default:
		close(node.subscriptions[n])
	}

	someSubscribed := false
	for _, s := range node.subscriptions {
		select {
		case <-s:
		default: // if at least one subscription is still open, we don't have to unsubscribe from inputs
			someSubscribed = true
		}
	}

	if !someSubscribed {
		close(node.allUnsubscribed)
		for _, input := range node.inputs {
			input.unsubscribe()
		}
	}
}

func loopOverChannel[C any, T any, R any](node workerNode[T, R], channel <-chan C, function func(value C) bool) {
	for {
		select { // the nested select gives priority to the quit signal, so we always exit early if needed
		case <-node.QuitSignal():
			return
		default:
			select {
			case <-node.QuitSignal():
				return
			case value, open := <-channel:
				if !open || !function(value) {
					return
				}
			}
		}
	}
}

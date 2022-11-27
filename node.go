package jpipe

import (
	"fmt"
	"sync"

	"github.com/junitechnology/jpipe/options"
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

	worker     worker[T, R]
	quitSignal chan struct{}
	doneSignal chan struct{}
}

type workerNode[T any, R any] interface {
	Inputs() []*Channel[T]
	LoopInput(i int, function func(value T) bool)
	Send(value R) bool
	QuitSignal() <-chan struct{}
	HandlePanic()
}

func newPipelineNode[T any, R any](
	nodeType string,
	pipeline *Pipeline,
	inputs []*Channel[T],
	numOutputs int,
	worker worker[T, R],
	opts ...options.NodeOption) (pipelineNode, []*Channel[R]) {

	buffered := getOptionOrDefault(opts, Buffered(0))

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
	}

	for _, input := range inputs {
		input.setToNode(node)
	}

	for i := 0; i < numOutputs; i++ {
		goChannel := make(chan R, buffered.Size)
		n := i // avoid capturing the loop var i in the unsubscriber closure
		node.outputs[i] = newChannel(pipeline, goChannel, func() { node.unsubscribe(n) })
		node.outputWriters[i] = goChannel
		node.subscriptions[i] = make(chan struct{})
	}

	pipeline.addNode(node)

	return node, node.outputs
}

func newLinearPipelineNode[T any, R any](nodeType string, input *Channel[T], worker worker[T, R], opts ...options.NodeOption) (pipelineNode, *Channel[R]) {
	node, outputs := newPipelineNode(nodeType, input.getPipeline(), []*Channel[T]{input}, 1, worker, opts...)
	return node, outputs[0]
}

func newSourcePipelineNode[R any](nodeType string, pipeline *Pipeline, worker worker[any, R], opts ...options.NodeOption) (pipelineNode, *Channel[R]) {
	node, outputs := newPipelineNode(nodeType, pipeline, []*Channel[any]{}, 1, worker, opts...)
	return node, outputs[0]
}

func newSinkPipelineNode[T any](nodeType string, input *Channel[T], worker worker[T, any], opts ...options.NodeOption) pipelineNode {
	node, _ := newPipelineNode(nodeType, input.getPipeline(), []*Channel[T]{input}, 0, worker, opts...)
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
				for _, input := range node.inputs {
					input.unsubscribe() // eagerly unsubscribe from inputs here
				}
			}
			close(node.quitSignal)
		}()

		defer node.HandlePanic()
		node.worker(node)
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

func (node *node[T, R]) HandlePanic() {
	if r := recover(); r != nil {
		node.pipeline.Cancel(fmt.Errorf("%v", r))
	}
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

	for _, s := range node.subscriptions {
		select {
		case <-s:
		default: // if at least one subscription is still open, we don't have to close allUnsubscribed
			return
		}
	}
	close(node.allUnsubscribed)
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

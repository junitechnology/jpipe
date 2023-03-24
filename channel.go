package jpipe

import "sync"

// A Channel is a wrapper for a Go channel.
// It provides chainable methods to construct pipelines, but conceptually it must be seen as a nothing but an enhanced Go channel.
type Channel[T any] struct {
	pipeline     *Pipeline
	channel      <-chan T
	unsubscriber func()
	toNode       pipelineNode
	mutex        sync.Mutex
}

func newChannel[T any](pipeline *Pipeline, channel chan T, unsubscriber func()) *Channel[T] {
	return &Channel[T]{
		pipeline:     pipeline,
		channel:      channel,
		unsubscriber: unsubscriber,
	}
}

func (c *Channel[T]) unsubscribe() {
	c.unsubscriber()
}

func (c *Channel[T]) setToNode(toNode pipelineNode) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.toNode != nil {
		panic("Can't subscribe more than one operator to a Channel")
	}

	c.toNode = toNode
}

func (c *Channel[T]) getToNode() pipelineNode {
	return c.toNode
}

func (c *Channel[T]) getPipeline() *Pipeline {
	return c.pipeline
}

func (c *Channel[T]) getChannel() <-chan T {
	return c.channel
}

func (c *Channel[T]) Pipeline() *Pipeline {
	return c.pipeline
}

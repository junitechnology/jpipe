package jpipe

// A Channel is a wrapper for a Go channel.
// It provides chainable methods to construct pipelines, but conceptually it must be seen as a nothing but an enhanced Go channel.
type Channel[T any] struct {
	pipeline     *Pipeline
	channel      <-chan T
	unsubscriber func()
	toNode       pipelineNode
}

func newChannel[T any](pipeline *Pipeline, channel chan T, unsubscriber func()) *Channel[T] {
	return &Channel[T]{
		pipeline:     pipeline,
		channel:      channel,
		unsubscriber: unsubscriber,
	}
}

func (p *Channel[T]) unsubscribe() {
	p.unsubscriber()
}

func (p *Channel[T]) setToNode(toNode pipelineNode) {
	p.toNode = toNode
}

func (p *Channel[T]) getToNode() pipelineNode {
	return p.toNode
}

func (p *Channel[T]) getPipeline() *Pipeline {
	return p.pipeline
}

func (p *Channel[T]) getChannel() <-chan T {
	return p.channel
}

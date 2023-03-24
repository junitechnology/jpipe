package jpipe

import (
	"context"
	"sync"
)

// A Pipeline is a container for the classic [Pipelines and Cancellation] pattern.
// Pipelines are safe to use from multiple goroutines.
//
// Pipeline works as a coordinator or context for multiple operators and Channels.
// It can also be seen as a sort of context and passed around as such.
//
// [Pipelines and Cancellation]: https://go.dev/blog/pipelines
type Pipeline struct {
	lock sync.Mutex

	context       context.Context
	startManually bool

	started bool
	done    chan struct{}
	err     error

	nodes       []pipelineNode
	activeNodes sync.WaitGroup
}

// A Config can be used to create a pipeline with certain settings
type Config struct {
	// Context is used by a Pipeline for cancellation.
	// If the context gets cancelled, the pipeline gets canceled too.
	Context context.Context
	// StartManually determines whether [Pipeline.Start] must be called manually.
	// If false, the first sink operator(ForEach, ToSlice, etc) to be created in the pipeline automatically starts it.
	// If true, the pipeline will be dormant until [Pipeline.Start] is called.
	StartManually bool
}

// New returns a pipeline with the given backing context.
// StartManually is false by default, meaning the pipeline will start when a sink operator(ForEach, ToSlice, etc) is created for it.
func New(ctx context.Context) *Pipeline {
	return NewPipeline(Config{Context: ctx})
}

// NewPipeline returns a Pipeline with the given [jpipe.Config]
func NewPipeline(config Config) *Pipeline {
	pipeline := Pipeline{
		done:          make(chan struct{}),
		startManually: config.StartManually,
	}

	if config.Context != nil {
		pipeline.context = config.Context
	} else {
		pipeline.context = context.TODO()
	}

	return &pipeline
}

// Start manually starts the pipeline.
// If the pipeline is already started, Start has no effect.
func (p *Pipeline) Start() {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.started {
		return
	}
	p.started = true

	go func() {
		select {
		case <-p.context.Done():
			p.Cancel(p.context.Err())
		case <-p.done:
		}
	}()

	for _, node := range p.nodes {
		node.Start()
		p.activeNodes.Add(1)
		n := node
		go func() {
			<-n.Done()
			p.activeNodes.Done()
		}()
	}
	go func() {
		p.activeNodes.Wait()
		p.Cancel(nil)
	}()
}

// Cancel manually cancels the pipeline with the given error
func (p *Pipeline) Cancel(err error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if !p.IsDone() { // no further action needed if the pipeline was already done
		if err != nil { // the if condition avoids a race condition when accessing Error()
			p.err = err
		}
		close(p.done)
	}
}

// Error returns the error in the pipeline if any.
// It returns nil if the pipeline is still running, or it completed successfully.
func (p *Pipeline) Error() error {
	return p.err
}

// Done returns a channel that's close when the pipeline either completed successfully or failed.
func (p *Pipeline) Done() <-chan struct{} {
	return p.done
}

// IsDone synchronously returns whether the pipeline completed successfully or failed.
func (p *Pipeline) IsDone() bool {
	select {
	case <-p.Done():
		return true
	default:
		return false
	}
}

func (p *Pipeline) addNode(node pipelineNode) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.started {
		node.Start() // nodes created for FlatMap after pipeline is started must be started immediately
		return
	}
	p.nodes = append(p.nodes, node)
	go func() { // done in a goroutine to avoid deadlock
		if node.IsSink() && !p.startManually {
			p.Start()
		}
	}()
}

func (p *Pipeline) Context() context.Context {
	return p.context
}

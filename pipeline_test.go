package jpipe_test

import (
	"context"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/stretchr/testify/assert"
)

func TestPipelineDoneWhenContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pipeline := jpipe.NewPipeline(jpipe.Config{Context: ctx})
	jpipe.
		FromSlice(pipeline, []int{1, 2, 3}).
		Interval(func(value int) time.Duration { return 100 * time.Millisecond }).
		ToSlice()

	select {
	case <-pipeline.Done():
		assert.Fail(t, "Pipeline must not be done initially")
	default:
	}
	assert.False(t, pipeline.IsDone())

	cancel()
	select {
	case <-pipeline.Done():
	case <-time.After(time.Millisecond):
		assert.Fail(t, "Pipeline must be done if context is canceled")
	}
	assert.True(t, pipeline.IsDone())
	assert.ErrorIs(t, pipeline.Error(), context.Canceled)
}

func TestPipelineRecoversFromPanicAndIncludesStacktrace(t *testing.T) {
	pipeline := jpipe.NewPipeline(jpipe.Config{})
	jpipe.
		FromSlice(pipeline, []int{1, 2, 3}).
		ForEach(func(value int) {
			panic("panic")
		})

	select {
	case <-pipeline.Done():
	case <-time.After(time.Millisecond):
		assert.Fail(t, "Pipeline must be done if context is canceled")
	}
	assert.True(t, pipeline.IsDone())
	t.Logf("panic error:\n %v", pipeline.Error().Error())
	assert.Contains(t, pipeline.Error().Error(), "panic")
	assert.Contains(t, pipeline.Error().Error(), "TestPipelineRecoversFromPanicAndIncludesStacktrace") // this shows that the stacktrace is included
}

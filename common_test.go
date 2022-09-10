package jpipe_test

import (
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/stretchr/testify/assert"
)

func drainChannel[T any](channel *jpipe.Channel[T]) []T {
	slice := []T{}
	for value := range channel.ToGoChannel() {
		slice = append(slice, value)
	}

	return slice
}

func readGoChannel[T any](channel <-chan T, n int) []T {
	slice := []T{}
	goChannel := channel
	for i := 0; i < n; i++ {
		slice = append(slice, <-goChannel)
	}

	return slice
}

func cancelPipeline(pipeline *jpipe.Pipeline) {
	pipeline.Cancel(nil)
	time.Sleep(time.Millisecond) // give some time for the cancellation to propagate
}

func assertChannelClosed[T any](t *testing.T, channel <-chan T) {
	select {
	case _, open := <-channel:
		assert.False(t, open, "Channel should be closed but a value was received")
	default:
		assert.Fail(t, "Channel should be closed")
	}
}

func assertChannelOpenButNoValue[T any](t *testing.T, channel <-chan T) {
	select {
	case _, open := <-channel:
		assert.False(t, open, "Channel should be open but a value should not be received")
	default:
	}
}

func assertPipelineDone(t *testing.T, pipeline *jpipe.Pipeline, timeout time.Duration) {
	select {
	case <-pipeline.Done():
	case <-time.After(timeout):
		assert.Fail(t, "Pipeline should be done")
	}
}

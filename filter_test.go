package jpipe_test

import (
	"context"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	t.Run("Filters values", func(t *testing.T) {
		pipeline := jpipe.New(context.Background())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3}).
			Filter(func(i int) bool { return i%2 == 1 })

		filteredValues := drainChannel(channel)

		assert.Equal(t, []int{1, 3}, filteredValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5}).
			Filter(func(i int) bool { return i%2 == 1 })
		goChannel := channel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestSkip(t *testing.T) {
	t.Run("Skips n values", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5}).Skip(2)

		values := drainChannel(channel)

		assert.Equal(t, []int{3, 4, 5}, values)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5}).Skip(2)
		goChannel := channel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestTake(t *testing.T) {
	t.Run("Takes n values", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}).Take(3)

		values := drainChannel(channel)

		assert.Equal(t, []int{1, 2, 3}, values)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}).Take(3)
		goChannel := channel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestDistinct(t *testing.T) {
	t.Run("Filters distinct values", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 1, 1, 2, 3, 4, 3})
		distinctChannel := jpipe.Distinct(channel, func(i int) int { return i })

		values := drainChannel(distinctChannel)

		assert.Equal(t, []int{1, 2, 3, 4}, values)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 1, 1, 2, 3, 4, 3})
		distinctChannel := jpipe.Distinct(channel, func(i int) int { return i })
		goChannel := distinctChannel.ToGoChannel()

		readGoChannel(goChannel, 3)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

package jpipe_test

import (
	"context"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
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

		assertChannelClosed(t, goChannel, 10*time.Millisecond)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Concurrency yields reduced processing times", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		start := time.Now()
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5}).
			Filter(func(i int) bool {
				time.Sleep(100 * time.Millisecond)
				return i%2 == 1
			}, jpipe.Concurrent(3))

		mappedValues := drainChannel(channel)
		elapsed := time.Since(start)

		slices.Sort(mappedValues) // The output order with concurrency is unpredictable
		assert.Equal(t, []int{1, 3, 5}, mappedValues)
		assert.Less(t, elapsed, 300*time.Millisecond) // It would have taken 500ms serially, but it takes about 200ms with 5 elements and concurrency 3
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Ordered concurrency yields reduced processing times", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		start := time.Now()
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5}).
			Filter(func(i int) bool {
				time.Sleep(100 * time.Millisecond)
				return i%2 == 1
			}, jpipe.Concurrent(3), jpipe.Ordered(3))

		mappedValues := drainChannel(channel)
		elapsed := time.Since(start)

		assert.Equal(t, []int{1, 3, 5}, mappedValues)
		assert.Less(t, elapsed, 300*time.Millisecond) // It would have taken 500ms serially, but it takes about 200ms with 5 elements and concurrency 3
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Ordered concurrency keeps input order on output", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromRange(pipeline, 1, 1000).
			Filter(func(i int) bool {
				time.Sleep(10 * time.Millisecond)
				return i%2 == 1
			}, jpipe.Concurrent(20), jpipe.Ordered(20))
		filteredValues := drainChannel(channel)

		assert.Len(t, filteredValues, 500)
		for i := 0; i < 500; i++ {
			assert.Equal(t, 2*i+1, filteredValues[i])
		}
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

		assertChannelClosed(t, goChannel, 10*time.Millisecond)
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

		assertChannelClosed(t, goChannel, 10*time.Millisecond)
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

		assertChannelClosed(t, goChannel, 10*time.Millisecond)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

package jpipe_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/junitechnology/jpipe/options"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestMap(t *testing.T) {
	t.Run("Maps values", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		mappedChannel := jpipe.Map(channel, func(i int) string { return fmt.Sprintf("%dA", i) })

		mappedValues := drainChannel(mappedChannel)

		assert.Equal(t, []string{"1A", "2A", "3A"}, mappedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if context done", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		mappedChannel := jpipe.Map(channel, func(i int) string { return fmt.Sprintf("%dA", i) })
		goChannel := mappedChannel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Concurrency yields reduced processing times", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5})
		start := time.Now()
		mappedChannel := jpipe.Map(channel, func(i int) string {
			time.Sleep(10 * time.Millisecond)
			return fmt.Sprintf("%dA", i)
		}, options.MapOptions{Concurrency: 3})

		mappedValues := drainChannel(mappedChannel)
		elapsed := time.Since(start)

		slices.Sort(mappedValues) // The output order with concurrency is unpredictable
		assert.Equal(t, []string{"1A", "2A", "3A", "4A", "5A"}, mappedValues)
		assert.Less(t, elapsed, 30*time.Millisecond) // It would have taken 50ms serially, but it takes about 20ms with 5 elements and concurrency 3
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestFlatMap(t *testing.T) {
	t.Run("FlatMaps values", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		mappedChannel := jpipe.FlatMap(channel, func(i int) *jpipe.Channel[string] {
			return jpipe.FromSlice(pipeline, []string{fmt.Sprintf("%dA", i), fmt.Sprintf("%dB", i)})
		})

		mappedValues := drainChannel(mappedChannel)

		assert.Equal(t, []string{"1A", "1B", "2A", "2B", "3A", "3B"}, mappedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if context done", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		mappedChannel := jpipe.FlatMap(channel, func(i int) *jpipe.Channel[string] {
			return jpipe.FromSlice(pipeline, []string{fmt.Sprintf("%dA", i), fmt.Sprintf("%dB", i)})
		})
		goChannel := mappedChannel.ToGoChannel()

		readGoChannel(goChannel, 4)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Concurrency yields reduced processing times", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5})
		start := time.Now()
		mappedChannel := jpipe.FlatMap(channel, func(i int) *jpipe.Channel[string] {
			time.Sleep(10 * time.Millisecond)
			return jpipe.FromSlice(pipeline, []string{fmt.Sprintf("%dA", i), fmt.Sprintf("%dB", i)})
		}, options.FlatMapOptions{Concurrency: 3})

		mappedValues := drainChannel(mappedChannel)
		elapsed := time.Since(start)

		slices.Sort(mappedValues) // The output order with concurrency is unpredictable
		assert.Equal(t, []string{"1A", "1B", "2A", "2B", "3A", "3B", "4A", "4B", "5A", "5B"}, mappedValues)
		assert.Less(t, elapsed, 30*time.Millisecond) // It would have taken 50ms serially, but it takes about 20ms with 5 elements and concurrency 3
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestBatch(t *testing.T) {
	t.Run("Batches values based on size only", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		batchedChannel := jpipe.Batch(channel, options.BatchOptions{Size: 3})

		batchedValues := drainChannel(batchedChannel)

		assert.Equal(t, [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}}, batchedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Batches values based on time only", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		sourceGoChannel := make(chan int)
		channel := jpipe.FromGoChannel(pipeline, sourceGoChannel)
		batchedChannel := jpipe.Batch(channel, options.BatchOptions{Timeout: 100 * time.Millisecond})

		time.AfterFunc(0, func() { sourceGoChannel <- 1 })
		time.AfterFunc(40*time.Millisecond, func() { sourceGoChannel <- 2 })
		time.AfterFunc(60*time.Millisecond, func() { sourceGoChannel <- 3 })
		time.AfterFunc(150*time.Millisecond, func() { sourceGoChannel <- 4 })
		time.AfterFunc(350*time.Millisecond, func() { sourceGoChannel <- 5 })
		time.AfterFunc(380*time.Millisecond, func() { close(sourceGoChannel) })
		batchedValues := drainChannel(batchedChannel)

		assert.Equal(t, [][]int{{1, 2, 3}, {4}, {}, {5}}, batchedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Batches values based on size and time", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		sourceGoChannel := make(chan int)
		channel := jpipe.FromGoChannel(pipeline, sourceGoChannel)
		batchedChannel := jpipe.Batch(channel, options.BatchOptions{Size: 2, Timeout: 100 * time.Millisecond})

		time.AfterFunc(0, func() { sourceGoChannel <- 1 })
		time.AfterFunc(40*time.Millisecond, func() { sourceGoChannel <- 2 })
		time.AfterFunc(60*time.Millisecond, func() { sourceGoChannel <- 3 })
		time.AfterFunc(280*time.Millisecond, func() { sourceGoChannel <- 4 })
		time.AfterFunc(300*time.Millisecond, func() { close(sourceGoChannel) })
		batchedValues := drainChannel(batchedChannel)

		assert.Equal(t, [][]int{{1, 2}, {3}, {}, {4}}, batchedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		sourceGoChannel := make(chan int, 5)
		channel := jpipe.FromGoChannel(pipeline, sourceGoChannel)
		batchedChannel := jpipe.Batch(channel, options.BatchOptions{Size: 2, Timeout: 100 * time.Millisecond})

		time.AfterFunc(0, func() { sourceGoChannel <- 1 })
		time.AfterFunc(40*time.Millisecond, func() { sourceGoChannel <- 2 })
		time.AfterFunc(60*time.Millisecond, func() { sourceGoChannel <- 3 })
		time.AfterFunc(180*time.Millisecond, func() { cancelPipeline(pipeline) })
		time.AfterFunc(280*time.Millisecond, func() { sourceGoChannel <- 4 })
		time.AfterFunc(300*time.Millisecond, func() { close(sourceGoChannel) })
		batchedValues := drainChannel(batchedChannel)

		assert.Equal(t, [][]int{{1, 2}, {3}}, batchedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

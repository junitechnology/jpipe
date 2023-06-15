package jpipe_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/junitechnology/jpipe/item"
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

		assertChannelClosed(t, goChannel, 10*time.Millisecond)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Concurrency yields reduced processing times", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5})
		start := time.Now()
		mappedChannel := jpipe.Map(channel, func(i int) string {
			time.Sleep(100 * time.Millisecond)
			return fmt.Sprintf("%dA", i)
		}, jpipe.Concurrent(3))

		mappedValues := drainChannel(mappedChannel)
		elapsed := time.Since(start)

		slices.Sort(mappedValues) // The output order with concurrency is unpredictable
		assert.Equal(t, []string{"1A", "2A", "3A", "4A", "5A"}, mappedValues)
		assert.Less(t, elapsed, 300*time.Millisecond) // It would have taken 500ms serially, but it takes about 200ms with 5 elements and concurrency 3
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Ordered concurrency yields reduced processing times", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5})
		start := time.Now()
		mappedChannel := jpipe.Map(channel, func(i int) string {
			time.Sleep(100 * time.Millisecond)
			return fmt.Sprintf("%dA", i)
		}, jpipe.Concurrent(3), jpipe.Ordered(3))

		mappedValues := drainChannel(mappedChannel)
		elapsed := time.Since(start)

		assert.Equal(t, []string{"1A", "2A", "3A", "4A", "5A"}, mappedValues)
		assert.Less(t, elapsed, 300*time.Millisecond) // It would have taken 500ms serially, but it takes about 200ms with 5 elements and concurrency 3
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Ordered concurrency keeps input order on output", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromRange(pipeline, 1, 1000)
		mappedChannel := jpipe.Map(channel, func(i int) string {
			time.Sleep(10 * time.Millisecond)
			return fmt.Sprintf("%dA", i)
		}, jpipe.Concurrent(20), jpipe.Ordered(20))

		mappedValues := drainChannel(mappedChannel)

		assert.Len(t, mappedValues, 1000)
		for i := 0; i < 1000; i++ {
			assert.Equal(t, fmt.Sprintf("%dA", i+1), mappedValues[i])
		}
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

		assertChannelClosed(t, goChannel, 10*time.Millisecond)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Concurrency yields reduced processing times", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5})
		start := time.Now()
		mappedChannel := jpipe.FlatMap(channel, func(i int) *jpipe.Channel[string] {
			time.Sleep(100 * time.Millisecond)
			return jpipe.FromSlice(pipeline, []string{fmt.Sprintf("%dA", i), fmt.Sprintf("%dB", i)})
		}, jpipe.Concurrent(3))

		mappedValues := drainChannel(mappedChannel)
		elapsed := time.Since(start)

		slices.Sort(mappedValues) // The output order with concurrency is unpredictable
		assert.Equal(t, []string{"1A", "1B", "2A", "2B", "3A", "3B", "4A", "4B", "5A", "5B"}, mappedValues)
		assert.Less(t, elapsed, 300*time.Millisecond) // It would have taken 500ms serially, but it takes about 200ms with 5 elements and concurrency 3
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

// TODO: Improve this test, the setup is awkward
func TestBatch(t *testing.T) {
	t.Run("Batches values based on size only", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		batchedChannel := jpipe.Batch(channel, 3, 0)

		batchedValues := drainChannel(batchedChannel)

		assert.Equal(t, [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}}, batchedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Batches values based on time only", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		sourceGoChannel := make(chan int)
		channel := jpipe.FromGoChannel(pipeline, sourceGoChannel)
		batchedChannel := jpipe.Batch(channel, 0, 100*time.Millisecond)

		go func() {
			sourceGoChannel <- 1
			<-time.After(40 * time.Millisecond) // 40
			sourceGoChannel <- 2
			<-time.After(20 * time.Millisecond) // 60
			sourceGoChannel <- 3
			<-time.After(90 * time.Millisecond) // 150
			sourceGoChannel <- 4
			<-time.After(200 * time.Millisecond) // 350
			sourceGoChannel <- 5
			<-time.After(30 * time.Millisecond) // 380
			close(sourceGoChannel)
		}()
		batchedValues := drainChannel(batchedChannel)

		assert.Equal(t, [][]int{{1, 2, 3}, {4}, {}, {5}}, batchedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Batches values based on size and time", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		sourceGoChannel := make(chan int)
		channel := jpipe.FromGoChannel(pipeline, sourceGoChannel)
		batchedChannel := jpipe.Batch(channel, 2, 100*time.Millisecond)

		go func() {
			sourceGoChannel <- 1
			<-time.After(40 * time.Millisecond) // 40
			sourceGoChannel <- 2
			<-time.After(20 * time.Millisecond) // 60
			sourceGoChannel <- 3
			<-time.After(220 * time.Millisecond) // 280
			sourceGoChannel <- 4
			<-time.After(20 * time.Millisecond) // 300
			close(sourceGoChannel)
		}()
		batchedValues := drainChannel(batchedChannel)

		assert.Equal(t, [][]int{{1, 2}, {3}, {}, {4}}, batchedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		sourceGoChannel := make(chan int, 5)
		channel := jpipe.FromGoChannel(pipeline, sourceGoChannel)
		batchedChannel := jpipe.Batch(channel, 2, 100*time.Millisecond)

		go func() {
			sourceGoChannel <- 1
			<-time.After(40 * time.Millisecond) // 40
			sourceGoChannel <- 2
			<-time.After(20 * time.Millisecond) // 60
			sourceGoChannel <- 3
			<-time.After(120 * time.Millisecond) // 180
			cancelPipeline(pipeline)
			<-time.After(100 * time.Millisecond) // 280
			sourceGoChannel <- 4
			<-time.After(20 * time.Millisecond) // 300
			close(sourceGoChannel)
		}()
		batchedValues := drainChannel(batchedChannel)

		assert.Equal(t, [][]int{{1, 2}, {3}}, batchedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestWrap(t *testing.T) {
	t.Run("Filters values", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		wrappedChannel := jpipe.Wrap(channel)

		values := drainChannel(wrappedChannel)

		assert.Equal(t, []item.Item[int]{{Value: 1}, {Value: 2}, {Value: 3}}, values)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		wrappedChannel := jpipe.Wrap(channel)
		goChannel := wrappedChannel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel, 10*time.Millisecond)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

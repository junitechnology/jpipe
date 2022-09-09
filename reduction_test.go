package jpipe_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/junitechnology/jpipe"
)

func TestReduce(t *testing.T) {
	t.Run("Reduces values", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4})
		reducedChannel := jpipe.Reduce(channel, func(state int64, value int) (int64, int64) {
			sum := state + int64(value)
			return sum, sum
		})

		mappedValues := drainChannel(reducedChannel)

		assert.Equal(t, []int64{1, 3, 6, 10}, mappedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if context done", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4})
		reducedChannel := jpipe.Reduce(channel, func(state int64, value int) (int64, int64) {
			sum := state + int64(value)
			return sum, sum
		})
		goChannel := reducedChannel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestMaxByReducer(t *testing.T) {
	t.Run("Returns maximum values by key keeping first", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []string{"C-1", "B-2", "C-1", "A-3", "D-3"})
		maxChannel := jpipe.MaxBy(channel, func(value string) string { return strings.Split(value, "-")[1] })

		mappedValues := drainChannel(maxChannel)

		assert.Equal(t, []string{"C-1", "B-2", "B-2", "A-3", "A-3"}, mappedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Returns maximum values by key keeping last", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []string{"C-1", "B-2", "C-1", "A-3", "D-3"})
		maxChannel := jpipe.MaxBy(channel, func(value string) string { return strings.Split(value, "-")[1] }, jpipe.MaxByOptions{Keep: jpipe.KEEP_LAST})

		mappedValues := drainChannel(maxChannel)

		assert.Equal(t, []string{"C-1", "B-2", "B-2", "A-3", "D-3"}, mappedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if context done", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []string{"C-1", "B-2", "C-1", "A-3"})
		minChannel := jpipe.MaxBy(channel, func(value string) string { return strings.Split(value, "-")[1] })
		goChannel := minChannel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestMinByReducer(t *testing.T) {
	t.Run("Returns minimum values by key keeping first", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []string{"A-3", "B-2", "A-3", "C-1", "D-1"})
		minChannel := jpipe.MinBy(channel, func(value string) string { return strings.Split(value, "-")[1] })

		mappedValues := drainChannel(minChannel)

		assert.Equal(t, []string{"A-3", "B-2", "B-2", "C-1", "C-1"}, mappedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Returns minimum values by key keeping last", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []string{"A-3", "B-2", "A-3", "C-1", "D-1"})
		minChannel := jpipe.MinBy(channel, func(value string) string { return strings.Split(value, "-")[1] }, jpipe.MinByOptions{Keep: jpipe.KEEP_LAST})

		mappedValues := drainChannel(minChannel)

		assert.Equal(t, []string{"A-3", "B-2", "B-2", "C-1", "D-1"}, mappedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if context done", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []string{"C-1", "B-2", "C-1", "A-3"})
		minChannel := jpipe.MinBy(channel, func(value string) string { return strings.Split(value, "-")[1] })
		goChannel := minChannel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestCount(t *testing.T) {
	t.Run("Returns maximum values by key", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{11, 12, 13, 14}).Count()

		mappedValues := drainChannel(channel)

		assert.Equal(t, []int64{1, 2, 3, 4}, mappedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if context done", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{11, 12, 13, 14}).Count()
		goChannel := channel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

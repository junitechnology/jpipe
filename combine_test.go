package jpipe_test

import (
	"context"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestMerge(t *testing.T) {
	t.Run("Merges values from all channels", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel1 := make(chan int)
		channel2 := make(chan int)
		channel3 := make(chan int)
		mergedChannel := jpipe.Merge(
			jpipe.FromGoChannel(pipeline, channel1),
			jpipe.FromGoChannel(pipeline, channel2),
			jpipe.FromGoChannel(pipeline, channel3))

		go func() {
			defer close(channel1)
			defer close(channel2)
			defer close(channel3)
			channel1 <- 1
			time.Sleep(time.Millisecond)
			channel2 <- 2
			time.Sleep(time.Millisecond)
			channel3 <- 3
			time.Sleep(time.Millisecond)
			channel1 <- 4
			time.Sleep(time.Millisecond)
			channel1 <- 5
			time.Sleep(time.Millisecond)
			channel2 <- 6
			time.Sleep(time.Millisecond)
			channel3 <- 7
			time.Sleep(time.Millisecond)
			channel2 <- 8
			time.Sleep(time.Millisecond)
			channel1 <- 9
		}()

		mergedValues := drainChannel(mergedChannel)
		slices.Sort(mergedValues)

		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, mergedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early on pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel1 := jpipe.FromSlice(pipeline, []int{1, 2, 5})
		channel2 := jpipe.FromSlice(pipeline, []int{3, 4, 6})
		mergedChannel := jpipe.Merge(channel1, channel2)
		goChannel := mergedChannel.ToGoChannel()

		readGoChannel(goChannel, 3)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel, 10*time.Millisecond)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestConcat(t *testing.T) {
	t.Run("Concats values from all channels", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel1 := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		channel2 := jpipe.FromSlice(pipeline, []int{4, 5, 6})
		channel3 := jpipe.FromSlice(pipeline, []int{7, 8, 9})
		mergedChannel := jpipe.Concat(channel1, channel2, channel3)

		mergedValues := drainChannel(mergedChannel)

		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, mergedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early on pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel1 := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		channel2 := jpipe.FromSlice(pipeline, []int{4, 5, 6})
		mergedChannel := jpipe.Concat(channel1, channel2)
		goChannel := mergedChannel.ToGoChannel()

		readGoChannel(goChannel, 4)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel, 10*time.Millisecond)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

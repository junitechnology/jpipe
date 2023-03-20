package jpipe_test

import (
	"sync"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestSplit(t *testing.T) {
	t.Run("Sends every value to any channel", func(t *testing.T) {
		pipeline := jpipe.NewPipeline(jpipe.Config{StartManually: true})
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		splitChannels := channel.Split(2)
		goChannel1 := splitChannels[0].ToGoChannel()
		goChannel2 := splitChannels[1].ToGoChannel()
		pipeline.Start()

		goChannels := []<-chan int{goChannel1, goChannel2}
		outputValues := make([][]int, 2)
		var wg sync.WaitGroup
		for i := range goChannels {
			wg.Add(1)
			idx := i
			go func() {
				defer wg.Done()
				for value := range goChannels[idx] {
					time.Sleep(time.Millisecond) // ensure the other channel gets values too
					outputValues[idx] = append(outputValues[idx], value)
				}
			}()
		}
		wg.Wait()

		allValues := []int{}
		for _, o := range outputValues {
			assert.NotEmpty(t, o)
			allValues = append(allValues, o...)
		}
		slices.Sort(allValues)
		assert.Equal(t, []int{1, 2, 3}, allValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Split exits early if context done", func(t *testing.T) {
		pipeline := jpipe.NewPipeline(jpipe.Config{StartManually: true})
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		splitChannels := channel.Split(2)
		goChannel1 := splitChannels[0].ToGoChannel()
		goChannel2 := splitChannels[1].ToGoChannel()
		pipeline.Start()

		<-goChannel1
		<-goChannel2
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel1, 10*time.Millisecond)
		assertChannelClosed(t, goChannel2, 10*time.Millisecond)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestBroadcast(t *testing.T) {
	t.Run("Broadcasts all values to each channel", func(t *testing.T) {
		pipeline := jpipe.NewPipeline(jpipe.Config{StartManually: true})
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		broadcastChannels := channel.Broadcast(2)
		goChannel1 := broadcastChannels[0].ToGoChannel()
		goChannel2 := broadcastChannels[1].ToGoChannel()
		pipeline.Start()

		assert.Equal(t, 1, <-goChannel1)
		assert.Equal(t, 1, <-goChannel2)
		assert.Equal(t, 2, <-goChannel1)
		assert.Equal(t, 2, <-goChannel2)
		assert.Equal(t, 3, <-goChannel1)
		assert.Equal(t, 3, <-goChannel2)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Broadcasts exits early if context done", func(t *testing.T) {
		pipeline := jpipe.NewPipeline(jpipe.Config{StartManually: true})
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})
		broadcastChannels := channel.Broadcast(2)
		goChannel1 := broadcastChannels[0].ToGoChannel()
		goChannel2 := broadcastChannels[1].ToGoChannel()
		pipeline.Start()

		<-goChannel1
		<-goChannel2
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel1, 10*time.Millisecond)
		assertChannelClosed(t, goChannel2, 10*time.Millisecond)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

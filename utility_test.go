package jpipe_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/stretchr/testify/assert"
)

func TestBuffer(t *testing.T) {
	t.Run("Buffers and unblocks sender", func(t *testing.T) {
		goChannel := make(chan int)
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromGoChannel(pipeline, goChannel).Buffer(3)

		var outputValues []int
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			outputValues = drainChannel(channel)
			wg.Done()
		}()

		for i := 1; i <= 3; i++ {
			select {
			case goChannel <- i:
			case <-time.After(10 * time.Millisecond):
				assert.Fail(t, "sender should not block if buffer is not yet full")
			}
		}
		go func() {
			for i := 4; i <= 5; i++ {
				goChannel <- i
			}
			close(goChannel)
		}()
		wg.Wait()

		assert.Equal(t, []int{1, 2, 3, 4, 5}, outputValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5}
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, slice).Buffer(3)

		values := []int{}
		<-channel.ForEach(func(i int) {
			values = append(values, i)
			if i == 3 {
				cancelPipeline(pipeline)
				time.Sleep(time.Millisecond) // some time for cancellation to propagate
			}
		})

		assert.Equal(t, []int{1, 2, 3}, values)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestTap(t *testing.T) {
	t.Run("Taps on the channel and applies function for each value", func(t *testing.T) {
		slice := []int{1, 2, 3}
		pipeline := jpipe.New(context.TODO())
		valuesFromTap := []int{}
		channel := jpipe.FromSlice(pipeline, slice).
			Tap(func(i int) { valuesFromTap = append(valuesFromTap, i) })

		outputValues := drainChannel(channel)

		assert.Equal(t, slice, valuesFromTap)
		assert.Equal(t, slice, outputValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		valuesFromTap := []int{}
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5, 6, 7}).
			Tap(func(i int) { valuesFromTap = append(valuesFromTap, i) })
		goChannel := channel.ToGoChannel()

		readGoChannel(goChannel, 2)
		time.Sleep(time.Millisecond) // give some time for the third value to pass the tap
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assert.Less(t, len(goChannel), 7)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestInterval(t *testing.T) {
	t.Run("Emits values with interval", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3}).
			Interval(func(value int) time.Duration { return 100 * time.Millisecond })

		values := []int{}
		goChannel := channel.ToGoChannel()
		values = append(values, <-goChannel)
		t0 := time.Now()
		values = append(values, <-goChannel)
		t1 := time.Now()
		values = append(values, <-goChannel)
		t2 := time.Now()

		assert.Equal(t, []int{1, 2, 3}, values)
		assert.WithinDuration(t, t0.Add(100*time.Millisecond), t1, 20*time.Millisecond)
		assert.WithinDuration(t, t1.Add(100*time.Millisecond), t2, 20*time.Millisecond)

		// We wait less than the interval. We want to assert it doesn't add an interval after the last value
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3}).
			Interval(func(value int) time.Duration { return 40 * time.Millisecond })
		goChannel := channel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 60*time.Millisecond)
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

		assertChannelClosed(t, goChannel1)
		assertChannelClosed(t, goChannel2)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

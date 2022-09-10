package jpipe_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestForEach(t *testing.T) {
	t.Run("Processes all values in the channel", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})

		values := []int{}
		<-channel.ForEach(func(value int) {
			values = append(values, value)
		})

		assert.Equal(t, []int{1, 2, 3}, values)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early on pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})

		values := []int{}
		<-channel.ForEach(func(i int) {
			if i == 2 {
				cancelPipeline(pipeline)
			}
			values = append(values, i)
		})

		assert.Equal(t, []int{1, 2}, values)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Concurrency yields reduced processing times", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

		values := []int{}
		lock := sync.Mutex{}
		start := time.Now()
		<-channel.ForEach(func(value int) {
			time.Sleep(10 * time.Millisecond)
			lock.Lock()
			values = append(values, value)
			lock.Unlock()
		}, jpipe.ForEachOptions{Concurrency: 5})
		elapsed := time.Since(start)

		slices.Sort(values) // The output order with concurrency is unpredictable
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, values)
		assert.Less(t, elapsed, 40*time.Millisecond) // It would have taken 100ms serially, but it takes about 20ms with 10 elements and concurrency 5
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestToSlice(t *testing.T) {
	slice := []int{1, 2, 3}
	pipeline := jpipe.New(context.TODO())
	channel := jpipe.FromSlice(pipeline, slice)

	actual := <-channel.ToSlice()

	assert.Equal(t, slice, actual)
	assertPipelineDone(t, pipeline, 10*time.Millisecond)
}

func TestToMap(t *testing.T) {
	t.Run("Converts to map keeping first values", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{11, 42, 31, 22, 73})

		actual := <-jpipe.ToMap(channel, func(i int) int { return i % 10 }, jpipe.ToMapOptions{Keep: jpipe.KEEP_FIRST})

		assert.Equal(t, map[int]int{1: 11, 2: 42, 3: 73}, actual)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Converts to map keeping last values", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{11, 42, 31, 22, 73})

		actual := <-jpipe.ToMap(channel, func(i int) int { return i % 10 }, jpipe.ToMapOptions{Keep: jpipe.KEEP_LAST})

		assert.Equal(t, map[int]int{1: 31, 2: 22, 3: 73}, actual)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestLast(t *testing.T) {
	t.Run("Returns last value", func(t *testing.T) {
		slice := []int{1, 2, 3}
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, slice)

		last := <-channel.Last()

		assert.Equal(t, 3, last)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Returned channel closes if no value seen", func(t *testing.T) {
		slice := []int{}
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, slice)

		_, open := <-channel.Last()

		assert.False(t, open)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestAny(t *testing.T) {
	t.Run("Returns true when any value passes the predicate", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		goChannel := make(chan int)
		channel := jpipe.FromGoChannel(pipeline, goChannel)

		resultChannel := channel.Any(func(i int) bool { return i > 3 })
		goChannel <- 1
		goChannel <- 2
		goChannel <- 3
		assertChannelOpenButNoValue(t, resultChannel)
		goChannel <- 4
		result := <-resultChannel

		assert.True(t, result)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Returns false if no value passed the predicate", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})

		result := <-channel.Any(func(i int) bool { return i > 3 })

		assert.False(t, result)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early on pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		goChannel := make(chan int)
		channel := jpipe.FromGoChannel(pipeline, goChannel)

		go func() {
			goChannel <- 1
			goChannel <- 2
			cancelPipeline(pipeline)
		}()

		result := <-channel.Any(func(i int) bool { return i > 3 })

		assert.False(t, result)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestAll(t *testing.T) {
	t.Run("Returns false when any value fails the predicate", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		goChannel := make(chan int)
		channel := jpipe.FromGoChannel(pipeline, goChannel)

		resultChannel := channel.All(func(i int) bool { return i <= 3 })
		goChannel <- 1
		goChannel <- 2
		goChannel <- 3
		assertChannelOpenButNoValue(t, resultChannel)
		goChannel <- 4
		result := <-resultChannel

		assert.False(t, result)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Returns true if all values passed the predicate", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})

		result := <-channel.Any(func(i int) bool { return i <= 3 })

		assert.True(t, result)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early on pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		goChannel := make(chan int)
		channel := jpipe.FromGoChannel(pipeline, goChannel)

		go func() {
			goChannel <- 1
			goChannel <- 2
			cancelPipeline(pipeline)
		}()

		result := <-channel.Any(func(i int) bool { return i <= 3 })

		assert.True(t, result)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestNone(t *testing.T) {
	t.Run("Returns false when some value passes the predicate", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		goChannel := make(chan int)
		channel := jpipe.FromGoChannel(pipeline, goChannel)

		resultChannel := channel.None(func(i int) bool { return i > 3 })
		goChannel <- 1
		goChannel <- 2
		goChannel <- 3
		assertChannelOpenButNoValue(t, resultChannel)
		goChannel <- 4
		result := <-resultChannel

		assert.False(t, result)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Returns true no value passed the predicate", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{1, 2, 3})

		result := <-channel.None(func(i int) bool { return i > 3 })

		assert.True(t, result)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early on pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		goChannel := make(chan int)
		channel := jpipe.FromGoChannel(pipeline, goChannel)

		go func() {
			goChannel <- 1
			goChannel <- 2
			cancelPipeline(pipeline)
		}()

		result := <-channel.None(func(i int) bool { return i > 3 })

		assert.True(t, result)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

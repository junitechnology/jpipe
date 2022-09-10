package jpipe_test

import (
	"context"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/stretchr/testify/assert"
)

func TestSumReducer(t *testing.T) {
	pipeline := jpipe.New(context.TODO())
	channel := jpipe.FromSlice(pipeline, []int{1, 2, 3, 4})
	mappedChannel := jpipe.Reduce(channel, jpipe.SumReducer[int, int64]())

	mappedValues := drainChannel(mappedChannel)

	assert.Equal(t, []int64{1, 3, 6, 10}, mappedValues)
	assertPipelineDone(t, pipeline, 10*time.Millisecond)
}

func TestMaxReducer(t *testing.T) {
	pipeline := jpipe.New(context.TODO())
	channel := jpipe.FromSlice(pipeline, []int{1, 2, 1, 3})
	mappedChannel := jpipe.Reduce(channel, jpipe.MaxReducer[int]())

	mappedValues := drainChannel(mappedChannel)

	assert.Equal(t, []int{1, 2, 2, 3}, mappedValues)
	assertPipelineDone(t, pipeline, 10*time.Millisecond)
}

func TestMinReducer(t *testing.T) {
	pipeline := jpipe.New(context.TODO())
	channel := jpipe.FromSlice(pipeline, []int{3, 2, 3, 1})
	mappedChannel := jpipe.Reduce(channel, jpipe.MinReducer[int]())

	mappedValues := drainChannel(mappedChannel)

	assert.Equal(t, []int{3, 2, 2, 1}, mappedValues)
	assertPipelineDone(t, pipeline, 10*time.Millisecond)
}

func TestAvgReducer(t *testing.T) {
	t.Run("Averages signed integers", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []int{-1, -2, 0, 1, 2})
		mappedChannel := jpipe.Reduce(channel, jpipe.AvgReducer[int, float64]())

		mappedValues := drainChannel(mappedChannel)

		assert.Equal(t, []float64{-1, -1.5, -1, -0.5, 0}, mappedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Averages unsigned integers", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []uint{1, 2, 3, 4})
		mappedChannel := jpipe.Reduce(channel, jpipe.AvgReducer[uint, float64]())

		mappedValues := drainChannel(mappedChannel)

		assert.Equal(t, []float64{1, 1.5, 2, 2.5}, mappedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Averages floats", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, []float32{1.5, 2.5, 3.5, 4.5})
		mappedChannel := jpipe.Reduce(channel, jpipe.AvgReducer[float32, float64]())

		mappedValues := drainChannel(mappedChannel)

		assert.Equal(t, []float64{1.5, 2, 2.5, 3}, mappedValues)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestCountReducer(t *testing.T) {
	pipeline := jpipe.New(context.TODO())
	channel := jpipe.FromSlice(pipeline, []int{11, 12, 13, 14})
	mappedChannel := jpipe.Reduce(channel, jpipe.CountReducer[int, int64]())

	mappedValues := drainChannel(mappedChannel)

	assert.Equal(t, []int64{1, 2, 3, 4}, mappedValues)
	assertPipelineDone(t, pipeline, 10*time.Millisecond)
}

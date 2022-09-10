package jpipe_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/junitechnology/jpipe"
	"github.com/stretchr/testify/assert"
)

func TestFromSlice(t *testing.T) {
	t.Run("Creates channel from slice", func(t *testing.T) {
		slice := []int{1, 2, 3}
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, slice)

		actual := drainChannel(channel)

		assert.Equal(t, slice, actual)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		slice := []int{1, 2, 3}
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromSlice(pipeline, slice)
		goChannel := channel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestFromRange(t *testing.T) {
	t.Run("Creates channel from range", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromRange(pipeline, 7, 9)
		actual := drainChannel(channel)
		assert.Equal(t, []int{7, 8, 9}, actual)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromRange(pipeline, 7, 9)
		goChannel := channel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

func TestFromGenerator(t *testing.T) {
	t.Run("Creates channel from generator", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromGenerator(pipeline, func(i uint64) string { return fmt.Sprintf("%dA", i) })
		goChannel := channel.ToGoChannel()

		actual := readGoChannel(goChannel, 3)

		assert.Equal(t, []string{"0A", "1A", "2A"}, actual)
	})

	t.Run("Exits early if pipeline canceled", func(t *testing.T) {
		pipeline := jpipe.New(context.TODO())
		channel := jpipe.FromGenerator(pipeline, func(i uint64) string { return fmt.Sprintf("%dA", i) })
		goChannel := channel.ToGoChannel()

		readGoChannel(goChannel, 2)
		cancelPipeline(pipeline)

		assertChannelClosed(t, goChannel)
		assertPipelineDone(t, pipeline, 10*time.Millisecond)
	})
}

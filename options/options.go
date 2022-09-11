package options

import "time"

type ForEachOptions struct {
	Concurrency int
}

type MapOptions struct {
	Concurrency int
	//Ordered bool
}

type FlatMapOptions struct {
	Concurrency int
	//Ordered bool
}

type BatchOptions struct {
	Size    int
	Timeout time.Duration
}

type BroadcastOptions struct {
	BufferSize int
}

type ToMapOptions struct {
	Keep KeepStrategy
}

type ReduceOptions[R any] struct {
	InitialState R
}

type KeepStrategy string

const (
	KEEP_FIRST KeepStrategy = "KEEP_FIRST"
	KEEP_LAST  KeepStrategy = "KEEP_LAST"
)

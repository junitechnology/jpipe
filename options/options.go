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

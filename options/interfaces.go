package options

type NodeOption interface {
	isNodeOption()
}

type PooledWorkerOption interface {
	isPooledWorkerOption()
}

type ForEachOption interface {
	isForEachOption()
}

type MapOption interface {
	isMapOption()
}

type FlatMapOption interface {
	isFlatMapOption()
}

type BroadcastOption interface {
	isBroadcastOption()
}

type ToMapOption interface {
	isToMapOption()
}

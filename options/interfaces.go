package options

type ForEachOptions interface {
	supportsForEach()
	//Concurrency int
}

type MapOptions interface {
	supportsMap()
	//Concurrency int
	//Ordered bool
}

type FlatMapOptions interface {
	supportsFlatMap()
	//Concurrency int
	//Ordered bool
}

type BroadcastOptions interface {
	supportsBroadcast()
	//BufferSize int
}

type ToMapOptions interface {
	supportsToMap()
	//Keep KeepStrategy
}

type KeepStrategy string

const (
	KEEP_FIRST KeepStrategy = "KEEP_FIRST"
	KEEP_LAST  KeepStrategy = "KEEP_LAST"
)

package options

type NodeOptions interface {
	supportsNode()
}

type ForEachOptions interface {
	supportsForEach()
}

type MapOptions interface {
	supportsMap()
}

type FlatMapOptions interface {
	supportsFlatMap()
}

type BroadcastOptions interface {
	supportsBroadcast()
}

type ToMapOptions interface {
	supportsToMap()
}

type KeepStrategy string

const (
	KEEP_FIRST KeepStrategy = "KEEP_FIRST"
	KEEP_LAST  KeepStrategy = "KEEP_LAST"
)

package options

type Concurrent struct {
	Concurrency int
}

func (c Concurrent) isPooledWorkerOption() {}
func (c Concurrent) isForEachOption()      {}
func (c Concurrent) isMapOption()          {}
func (c Concurrent) isFlatMapOption()      {}

type Ordered struct {
	OrderBufferSize int
}

func (c Ordered) isPooledWorkerOption() {}
func (c Ordered) isMapOption()          {}
func (c Ordered) isFlatMapOption()      {}

type Buffered struct {
	Size int
}

func (c Buffered) isNodeOption()      {}
func (b Buffered) isBroadcastOption() {}

type Keep struct {
	Strategy KeepStrategy
}

type KeepStrategy string

const (
	KEEP_FIRST KeepStrategy = "KEEP_FIRST"
	KEEP_LAST  KeepStrategy = "KEEP_LAST"
)

func (b Keep) isToMapOption() {}

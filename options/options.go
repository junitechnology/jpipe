package options

type Concurrent struct {
	Concurrency int
}

func (c Concurrent) isPooledWorkerOption() {}
func (c Concurrent) isForEachOption()      {}
func (c Concurrent) isMapOption()          {}
func (c Concurrent) isFlatMapOption()      {}
func (c Concurrent) isFilterOption()       {}
func (c Concurrent) isTapOption()          {}

type Ordered struct {
	OrderBufferSize int
}

func (o Ordered) isPooledWorkerOption() {}
func (o Ordered) isMapOption()          {}
func (o Ordered) isFlatMapOption()      {}
func (o Ordered) isFilterOption()       {}
func (o Ordered) isTapOption()          {}

type Buffered struct {
	Size int
}

func (b Buffered) isNodeOption()      {}
func (b Buffered) isSplitOption()     {}
func (b Buffered) isBroadcastOption() {}

type Keep struct {
	Strategy KeepStrategy
}

type KeepStrategy string

const (
	KEEP_FIRST KeepStrategy = "KEEP_FIRST"
	KEEP_LAST  KeepStrategy = "KEEP_LAST"
)

func (k Keep) isToMapOption() {}

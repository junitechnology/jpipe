package options

type Concurrent struct {
	Concurrency int
}

func (c Concurrent) supportsNode()    {}
func (c Concurrent) supportsForEach() {}
func (c Concurrent) supportsMap()     {}
func (c Concurrent) supportsFlatMap() {}

type Ordered struct {
	OrderBufferSize int
}

type Buffered struct {
	Size int
}

func (b Buffered) supportsNode()      {}
func (b Buffered) supportsBroadcast() {}

type Keep struct {
	Strategy KeepStrategy
}

func (b Keep) supportsToMap() {}

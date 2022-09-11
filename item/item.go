package item

import "context"

// An Item is a utility type to use in pipelines.
// It can have a Value or an Error, so it provides an easy way to propagate per-value(possibily recoverable) errors.
// It also carries a context that can be enriched along the pipeline.
type Item[T any] struct {
	Value T
	Error error
	Ctx   context.Context
}

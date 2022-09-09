package jpipe

import "context"

// An Item is a utility type to use in pipelines.
// It can have a Value or an Error, so it provides an easy way to propagate per-value(possibily recoverable) errors.
// It also carries a context that can be enriched along the pipeline.
type Item[T any] struct {
	Value T
	Error error
	Ctx   context.Context
}

type KeepStrategy string

const (
	KEEP_FIRST KeepStrategy = "KEEP_FIRST"
	KEEP_LAST  KeepStrategy = "KEEP_LAST"
)

func getOptions[T any](defaultOptions T, options []T) T {
	if len(options) > 0 {
		return options[0]
	}
	return defaultOptions
}

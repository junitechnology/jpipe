package jpipe

import (
	"context"

	"github.com/junitechnology/jpipe/item"
)

func Item[T any](value T, err error, ctx context.Context) item.Item[T] {
	return item.Item[T]{
		Value: value,
		Error: err,
		Ctx:   ctx,
	}
}

func ValueItem[T any](value T) item.Item[T] {
	return item.Item[T]{Value: value}
}

func ErrorItem[T any](err error) item.Item[T] {
	return item.Item[T]{Error: err}
}

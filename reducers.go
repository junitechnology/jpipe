package jpipe

import (
	"reflect"

	"golang.org/x/exp/constraints"
)

// Any number. Can be either a integer or a float
type Number interface {
	constraints.Integer | constraints.Float
}

// SimpleReducer creates a reducer from a simple reducer, like what most reduce implementations use.
// In a simple reducer, the state and the return value are the same.
func SimpleReducer[T any, R any](reducer func(R, T) R) func(R, T) (R, R) {
	return func(state R, value T) (R, R) {
		state = reducer(state, value)
		return state, state
	}
}

// SumReducer returns a reducer that sums all input values
func SumReducer[T Number, R Number]() func(R, T) (R, R) {
	return SimpleReducer(func(sum R, value T) R {
		return sum + R(value)
	})
}

// MaxReducer returns a reducer that returns the maximum value up to that point
func MaxReducer[T constraints.Ordered]() func(*T, T) (*T, T) {
	return func(max *T, value T) (*T, T) {
		if max == nil || value > *max {
			return &value, value
		}
		return max, *max
	}
}

// MinReducer returns a reducer that returns the minimum value up to that point
func MinReducer[T constraints.Ordered]() func(*T, T) (*T, T) {
	return func(min *T, value T) (*T, T) {
		if min == nil || value < *min {
			return &value, value
		}
		return min, *min
	}
}

type AvgState struct {
	Sum   any
	Count int64
}

// AvgReducer returns a reducer that returns the average of all values up to that point
func AvgReducer[T Number, R Number]() func(AvgState, T) (AvgState, R) {
	var t T
	reflectValue := reflect.ValueOf(t)
	if reflectValue.CanInt() {
		return func(state AvgState, value T) (AvgState, R) {
			if state.Sum == nil {
				state.Sum = int64(0)
			}
			state.Sum = state.Sum.(int64) + int64(value)
			state.Count++

			return state, R(state.Sum.(int64)) / R(state.Count)
		}
	} else if reflectValue.CanUint() {
		return func(state AvgState, value T) (AvgState, R) {
			if state.Sum == nil {
				state.Sum = uint64(0)
			}
			state.Sum = state.Sum.(uint64) + uint64(value)
			state.Count++

			return state, R(state.Sum.(uint64)) / R(state.Count)
		}
	} else {
		return func(state AvgState, value T) (AvgState, R) {
			if state.Sum == nil {
				state.Sum = float64(0)
			}
			state.Sum = state.Sum.(float64) + float64(value)
			state.Count++

			return state, R(state.Sum.(float64)) / R(state.Count)
		}
	}
}

// CountReducer returns a reducer that counts input elements
func CountReducer[T any, R constraints.Integer]() func(R, T) (R, R) {
	return SimpleReducer(func(count R, value T) R {
		return count + 1
	})
}

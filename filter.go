package jpipe

// Filter sends to the output channel only the input values that satisfy the predicate.
//
// ## Example:
//
//  output := input.Filter(func(i int) bool { return i%2==1 })
//
//  input : 0--1--2--3--4--5-X
//  output: ---1-----3-----5-X
func (input *Channel[T]) Filter(predicate func(T) bool) *Channel[T] {
	worker := func(node workerNode[T, T]) {
		node.LoopInput(0, func(value T) bool {
			if predicate(value) {
				return node.Send(value)
			}
			return true
		})
	}

	_, output := newLinearPipelineNode("Filter", input, 0, worker)
	return output
}

// Skip skips the first n input values, and then starts sending values from n+1 on to the output channel
//
// ## Example:
//
//  output := input.Skip(2)
//
//  input : 0--1--2--3--4--5-X
//  output: ------2--3-----5-X
func (input *Channel[T]) Skip(n uint64) *Channel[T] {
	worker := func(node workerNode[T, T]) {
		i := uint64(0)
		node.LoopInput(0, func(value T) bool {
			if i < n {
				i++
				return true
			}
			return node.Send(value)
		})
	}

	_, output := newLinearPipelineNode("Skip", input, 0, worker)
	return output
}

// Take sends the first n input values to the output channel, and then stops processing and closes the output channel.
//
// ## Example:
//
//  output := input.Take(3)
//
//  input : 0--1--2--3--4--5-X
//  output: 0--1--2-X
func (input *Channel[T]) Take(n uint64) *Channel[T] {
	worker := func(node workerNode[T, T]) {
		i := uint64(0)
		node.LoopInput(0, func(value T) bool {
			if i == n {
				return false
			}
			i++
			return node.Send(value)
		})
	}

	_, output := newLinearPipelineNode("Take", input, 0, worker)
	return output
}

// Distinct sends only input values it hasn't seen before to the output channel.
// It uses an internal map to keep track of all values seen,
// so keep in mind that it could exhaust memory if too many distinct values are received.
//
// ## Example:
//
//  output := Distinct(input, func(value int) int { return value })
//
//  input : 0--1--2--1--3--2-X
//  output: 0--1--2-----3----X
func Distinct[T any, K comparable](input *Channel[T], getKey func(T) K) *Channel[T] {
	worker := func(node workerNode[T, T]) {
		seen := map[any]bool{}
		node.LoopInput(0, func(value T) bool {
			key := getKey(value)
			if _, keySeen := seen[key]; keySeen {
				return true
			}
			seen[key] = true
			return node.Send(value)
		})
	}

	_, output := newLinearPipelineNode("Distinct", input, 0, worker)
	return output
}

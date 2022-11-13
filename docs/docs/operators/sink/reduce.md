---
layout: default
title: Reduce
parent: Sink
grand_parent: Operators
---

<h1>Reduce</h1>

```go
func Reduce[T any, R any](input *Channel[T], reducer func(R, T) R) <-chan R
```

`Reduce` performs a stateful reduction of the input values.
The reducer receives the current state and the current value, and must return the new state.

The final state is sent to the returned channel when all input values have been processed, or the pipeline is canceled.

<h2>Example</h2>
Calculating the sum of all input values
```go
output := Reduce(input, func(acc int64, value int) int64 { return acc + int64(value) })
```
![](/assets/images/diagrams/sink/reduce.svg){:class="img-responsive"}
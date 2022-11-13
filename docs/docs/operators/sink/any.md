---
layout: default
title: Any
parent: Sink
grand_parent: Operators
---

<h1>Any</h1>

```go
func (input *Channel[T]) Any(predicate func(T) bool) <-chan bool
```

`Any` determines if any input value matches the predicate.

If no value matches the predicate, `false` is sent to the returned channel when all input values have been processed, or the pipeline is canceled.

If instead some value is found to match the predicate, `true` is immediately sent to the returned channel and no more input values are read.

<h2>Examples</h2>

```go
output := input.Any(func(value int) bool { return value > 3 })
```
![](/assets/images/diagrams/sink/any-1.svg){:class="img-responsive"}

```go
output := input.Any(func(value int) bool { return value >= 2 })
```
![](/assets/images/diagrams/sink/any-2.svg){:class="img-responsive"}
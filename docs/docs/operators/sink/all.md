---
layout: default
title: All
parent: Sink
grand_parent: Operators
---

<h1>All</h1>

```go
func (input *Channel[T]) All(predicate func(T) bool) <-chan bool
```

`All` determines if all input values match the predicate.

If all values match the predicate, `true` is sent to the returned channel when all input values have been processed, or the pipeline is canceled.

If instead some value does not match the predicate, `false` is immediately sent to the returned channel and no more input values are read.

<h2>Examples</h2>

```go
output := input.All(func(value int) bool { return value < 4 })
```
![](/assets/images/diagrams/sink/all-1.svg){:class="img-responsive"}

```go
output := input.All(func(value int) bool { return value < 2 })
```
![](/assets/images/diagrams/sink/all-2.svg){:class="img-responsive"}
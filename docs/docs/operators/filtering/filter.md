---
layout: default
title: Filter
parent: Filtering
grand_parent: Operators
---

<h1>Filter</h1>

```go
func (input *Channel[T]) Filter(predicate func(T) bool) *Channel[T]
```

`Filter` sends to the output channel only the input values that match the predicate function.

<h2>Example</h2>

```go
output := input.Filter(func(x int) bool { return x%2 == 1 })
```
![](/assets/images/diagrams/filtering/filter.svg){:class="img-responsive"}
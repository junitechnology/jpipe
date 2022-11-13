---
layout: default
title: Interval
parent: Utility
grand_parent: Operators
---

<h1>Interval</h1>

```go
func (input *Channel[T]) Interval(interval func(value T) time.Duration) *Channel[T]
```

`Interval` transparently passes all input values to the output channel,
but a time interval is awaited after each element before sending another one.
No value is sent to the output while that interval is active.

This operator is prone to generating backpressure, so use it with care, and consider adding a `Buffer` before it.

<h2>Example</h2>

```go
output := input.Interval(4*time.Millisecond)
```
![](/assets/images/diagrams/utility/interval.svg){:class="img-responsive"}
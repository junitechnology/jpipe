---
layout: default
title: FromRange
parent: Source
grand_parent: Operators
---

<h1>FromRange</h1>

```go
func FromRange[T constraints.Integer](pipeline *Pipeline, start T, end T) *Channel[T]
```

`FromRange` creates a `Channel` from a range of integers.
All integers between `start` and `end` (both inclusive) are sent to the channel in order.

<h2>Example</h2>

```go
channel := FromSlice(pipeline, 1, 3)
```
![](/assets/images/diagrams/source/from-range.svg){:class="img-responsive"}
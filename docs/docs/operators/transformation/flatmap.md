---
layout: default
title: FlatMap
parent: Transformation
grand_parent: Operators
---

<h1>FlatMap</h1>

```go
func FlatMap[T any, R any](input *Channel[T], mapper func(T) *Channel[R], opts ...options.FlatMapOptions) *Channel[R]
```

`FlatMap` transforms every input value into a Channel and for each of those, it sends all values to the output channel.

<h2>Example</h2>

```go
output := FlatMap(input, func(i int) *Channel[int] { return FromSlice([]int{i, i * 10}) })
```
![](/assets/images/diagrams/transformation/flatmap.svg){:class="img-responsive"}
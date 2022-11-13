---
layout: default
title: Map
parent: Transformation
grand_parent: Operators
---

<h1>Map</h1>

```go
func Map[T any, R any](input *Channel[T], mapper func(T) R, opts ...options.MapOptions) *Channel[R]
```

`Map` transforms every input value with a mapper function and sends the results to the output channel.

<h2>Example</h2>

```go
output := Map(input, func(i int) int { return i * 10 })
```
![](/assets/images/diagrams/transformation/map.svg){:class="img-responsive"}
---
layout: default
title: ToMap
parent: Sink
grand_parent: Operators
---

<h1>ToMap</h1>

```go
func ToMap[T any, K comparable](input *Channel[T], getKey func(T) K, opts ...options.ToMapOptions) <-chan map[K]T
```

`ToMap` puts all values coming from the input channel in a map, using the `getKey` parameter to calculate the key.

The resulting map is sent to the returned channel when all input values have been processed, or the pipeline is canceled.

<h2>Example</h2>

```go
output := ToMap(input, func(value string) string { return strings.Split(value, "_")[0] })
```
![](/assets/images/diagrams/sink/to-map.svg){:class="img-responsive"}
```
result = {A:A_0, B:B_1, C:C_2}
```
---
layout: default
title: Distinct
parent: Filtering
grand_parent: Operators
---

<h1>Distinct</h1>

```go
func Distinct[T any, K comparable](input *Channel[T], getKey func(T) K) *Channel[T]
```

`Distinct` sends only input values for which the key hasn't been seen before to the output channel.
In simple words, it deduplicates the input channel.
It uses an internal map to keep track of all keys seen,
so keep in mind that it could exhaust memory if too many distinct values are received.

<h2>Example</h2>

```go
output := Distinct(input, func(value int) int { return value })
```
![](/assets/images/diagrams/filtering/distinct.svg){:class="img-responsive"}
---
layout: default
title: Tap
parent: Utility
grand_parent: Operators
---

<h1>Tap</h1>

```go
func (input *Channel[T]) Tap(function func(T), opts ...options.TapOption) *Channel[T]
```

`Tap` runs a function as a side effect for each input value,
and then sends the input values transparently to the output channel.
A common use case is logging.

<h2>Example</h2>

```go
output := input.Tap(func(x int) { fmt.Println(x)})
```
![](/assets/images/diagrams/utility/tap.svg){:class="img-responsive"}
```
Console output:
1
2
3
```
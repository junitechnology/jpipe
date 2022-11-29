---
layout: default
title: Split
parent: Fan-out
grand_parent: Operators
---

<h1>Split</h1>

```go
func (input *Channel[T]) Split(numOutputs int, opts ...options.SplitOption) []*Channel[T]
```

`Split` sends each input value to any of the output channels, with no specific priority.

<h2>Example</h2>

```go
outputs := input.Split(2)
```
![](/assets/images/diagrams/fanout/split.svg){:class="img-responsive"}
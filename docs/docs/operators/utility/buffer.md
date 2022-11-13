---
layout: default
title: Buffer
parent: Utility
grand_parent: Operators
---

<h1>Buffer</h1>

```go
func (input *Channel[T]) Buffer(n int) *Channel[T]
```

`Buffer` transparently passes input values to the output channel, but the output channel is buffered.
It is useful to avoid backpressure from slow consumers.

<h2>Example</h2>

```go
output := input.Buffer(2)
```
![](/assets/images/diagrams/utility/buffer.svg){:class="img-responsive"}
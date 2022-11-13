---
layout: default
title: Count
parent: Sink
grand_parent: Operators
---

<h1>Count</h1>

```go
func (input *Channel[T]) Count() <-chan int64
```

`Count` counts input values and sends the final count to the output channel.

The final count is sent to the return channel when all input values have been processed, or the pipeline is canceled.

<h2>Example</h2>

```go
output := input.Count()
```
![](/assets/images/diagrams/sink/count.svg){:class="img-responsive"}
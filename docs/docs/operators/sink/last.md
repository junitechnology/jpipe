---
layout: default
title: Last
parent: Sink
grand_parent: Operators
---

<h1>Last</h1>

```go
func (input *Channel[T]) Last() <-chan T
```

`Last` sends the last value received from the input channel to the output channel.

The last value is sent to the returned channel when all input values have been processed, or the pipeline is canceled.

<h2>Example</h2>

```go
output := input.Last()
```
![](/assets/images/diagrams/sink/last.svg){:class="img-responsive"}
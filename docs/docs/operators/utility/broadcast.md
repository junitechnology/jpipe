---
layout: default
title: Broadcast
parent: Utility
grand_parent: Operators
---

<h1>Broadcast</h1>

```go
func (input *Channel[T]) Broadcast(numOutputs int, opts ...options.BroadcastOptions) []*Channel[T]
```

`Broadcast` sends each input value to every output channel.
The next input value is not read by this operator until all output channels have read the current one.

Bear in mind that if one of the output channels is a slow consumer, it may block the other consumers.
This is a particularly annoying type of backpressure, cause not only does it block the input, it also blocks other consumers.
To avoid this, consider using `options.Buffered` and the output channels will be buffered, with no need for an extra `Buffer` operator.

<h2>Example</h2>

```go
outputs := input.Broadcast(2)
```
![](/assets/images/diagrams/utility/broadcast.svg){:class="img-responsive"}
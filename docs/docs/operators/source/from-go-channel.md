---
layout: default
title: FromGoChannel
parent: Source
grand_parent: Operators
---

<h1>FromGoChannel</h1>

```go
func FromGoChannel[T any](pipeline *Pipeline, channel <-chan T) *Channel[T]
```

`FromGoChannel` creates a `Channel` from a Go channel.

<h2>Example</h2>

```go
channel := FromGoChannel(pipeline, goChannel)
```
![](/assets/images/diagrams/source/from-go-channel.svg){:class="img-responsive"}
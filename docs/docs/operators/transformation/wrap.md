---
layout: default
title: Wrap
parent: Transformation
grand_parent: Operators
---

<h1>Wrap</h1>

```go
func Wrap[T any](input *Channel[T]) *Channel[item.Item[T]]
```

`Wrap` wraps every input value `T` in an `Item[T]` and sends it to the output channel.

`Item[T]` is used mostly to represent items that can have either a value or an error.
It's useful in situations where errors are an acceptable value in the pipeline,
and a single error does not represent a full pipeline error.

Another use for `Item[T]` is using the `Context` in it and enriching it in successive operators.

<h2>Example</h2>

```go
output := Wrap(input)
```
![](/assets/images/diagrams/transformation/wrap.svg){:class="img-responsive"}
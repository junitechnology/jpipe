---
layout: default
title: Operator Options
nav_order: 2
parent: Usage
---

<h1>Operator options</h1>

Some operators allow tuning their behaviour through options. Take e.g. `Map`:

```go
func Map[T any, R any](input *Channel[T], mapper func(T) R, opts ...options.MapOptions) *Channel[R]
```

The simplest way you can use `Map` is with no options:

```go
channel :=  jpipe.Map(channel, func(x int) int { return x * 10 })
```

So if you don't care about options, you can just pretend they don't exist, since they are a variadic parameter. But what is this `options.MapOptions` type? Just an interface:

```go
type MapOptions interface {
    supportsMap()
}
```

The interface is implemented by both `options.Concurrent` and `options.Ordered`, so you can use any of those.

The best way to create those is through utility functions in `jpipe` package: `jpipe.Concurrent(int)` and `jpipe.Ordered(int)`. But you don't need to know all of this for each operator, or even look for it. After you've passed the mandatory parameters to `Map`(`pipeline` and `mapper`), just type `jpipe.` will autosuggest you the available options. You end up with:

```go
channel :=  jpipe.Map(channel, func(x int) int { return x * 10 }, jpipe.Concurrent(2), jpipe.Ordered(10))
```

The full list of options is:

- `jpipe.Concurrent(concurrency int)`: Controls the concurrency of the operator.
- `jpipe.Ordered(bufferSize int)`: Makes the operator output ordered(same order as input).
- `jpipe.Buffered(size int)`: Makes the output channel(s) of the operator buffered.
- `jpipe.KeepFirst` and `jpipe.KeepLast`: If the operator must select a value out of many, this option controls whether it picks the first or the last one.

The actual usage of these options will become easier to understand as you progress through this guide.
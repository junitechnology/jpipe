---
layout: default
title: Chaining
nav_order: 3
parent: Usage
---

<h1>Chaining</h1>

The API tries to have chainable methods where possible. This allows compact and fluent processing pipelines like this:

```go
<-jpipe.FromRange(pipeline, 1, 1000).
    Filter(func(id int) bool { return !idExists(id) }).
    Take(5).
    ForEach(func(id int) { createEntityWithId(id) })
```

The snippet above goes over all integers from 1 to 1000, and for the first 5 that don't exist yet, it creates them. Nothing that a simple for loop wouldn't do in a clearer way, but this version can be made concurrent by just adding an option.

Now imagine we instead want to get the names of the first 5 entities that already exist:

```go
channel := jpipe.FromRange(pipeline, 1, 1000).
    Filter(func(id int) bool { return idExists(id) })
    Take(5).
names := <-jpipe.Map(channel, func(id int) string { return getNameForId(id) }).
    ToSlice()
```

You can see that `Map` "broke" our chain, because it's not a method on `Channel`. It's just a function that takes a `Channel` as a parameter. What's stopping us from making `Map` and other operators chainable, is a couple of limitations in Go generics:

<h2>Limitation 1: Methods must have no type parameters</h2>

If we tried to make `Map` a method like this:

```go
func (input *Channel[T]) Map[T any, R any](mapper func(T) R, opts ...options.MapOptions) *Channel[R]
```

We would be greeted by this syntax error: `method must have no type parameters`. This was a design decision in Go generics and the reasons for it can be found [here](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#no-parameterized-methods).

<h2>Limitation 2: Instantiation cycle (InvalidInstanceCycle)</h2>

Now let's try to make Batch chainable:

```go
func (input *Channel[T]) Batch[T any](size int, timeout time.Duration) *Channel[[]T]
```

This time we face a compiler error `instantiation cycle (InvalidInstanceCycle)`. You can find an explanation of the problem in [this GitHub issue](https://github.com/golang/go/issues/50215) from the Go repository.

The issue here is that Go instantiates types eagerly. So instantiating `Channel[T]`, with its `Batch` method, would also require to instantiate `Batch`'s return type `Channel[[]T]`. Doing so would in turn require to instantiate type `Channel[[][]T]`. And you can see how this would go on infinitely, resulting in an instantiation cycle.

<h2>Going forward</h2>

We can't have all operators be methods, but we do have the simple workaround of making them functions that take a `Channel` as parameter. The beauty of the API resents this, but usability-wise, this is good enough.

There's the probability that Go will evolve into solving these issues in future releases, and in that case, we would adapt our API accordingly. On a practical note though, both issues are apparently hard to solve, so we don't expect their solution any time soon, and it's not even clear if there's interest to solve them. That being said, we encourage you to use this and other libraries that make heavy use of generics, as that will be a motivation for the Go team to improve their implementation.

An interesting article that covers both limitations and how they impact the adoption of the functional programming paradigm in Go can be found [here](https://betterprogramming.pub/generics-in-go-are-we-there-yet-af851c35ba0).

If you also happen to think Go(and our library) would benefit from shorthand lambda expressions (arrow functions), give [this proposal](https://github.com/golang/go/issues/21498) some love.
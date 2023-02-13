---
layout: home
title: Overview
nav_order: 1
---

<h1>Overview</h1>

Jpipe is a library that implements the pipeline pattern for Go.
The pattern has been described by members of the core Go team several times in the past:

- [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
- [Google I/O 2012 - Go Concurrency Patterns](https://www.youtube.com/watch?v=f6kdp27TYZs)
- [Advanced Go Concurrency Patterns](https://go.dev/blog/io2013-talk-concurrency)

Go provides very powerful concurrency primitives, but implementing the pipeline pattern correctly,
with a correct handling of cancellation, requires a very good understanding of those primitives, and some non-negligible amount of boilerplate code. As pipelines become more complex, that boilerplate also starts to weigh heavily on code readability.

So let's see a very simple concurrency example. Let's assume we have an expensive IO operation that takes 1 second to execute:

```go
func expensiveIOOperation(id int) {
    time.Sleep(time.Second)
}
```

Imagine this operation must be run for ids 1 through 10. We don't want to wait 10 seconds though, so we decide to do it with a concurrency factor of 5, expecting to get the full operation down to 2 seconds. Let's see the full Go code for that:

<details>
<summary markdown="span">Plain Go version</summary>

```go
ids := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
channel := make(chan int)
concurrency := 5
var wg sync.WaitGroup
for i := 0; i < concurrency; i++ {
  wg.Add(1)
  go func() {
    defer wg.Done()
    for id := range channel {
      expensiveIOOperation(id)
    }
  }()
}

outer:
for _, id := range ids {
  select {
  // The nested select gives priority to the ctx.Done() signal, so we always exit early if needed
  // Without it, a single select just has no priority, so a new value could be processed even if the context has been canceled
  case <-ctx.Done():
    break outer
  default:
    select {
    case channel <- id:
    case <-ctx.Done(): // always check ctx.Done() to avoid leaking the goroutine
      break outer
    }
  }
}
close(channel)

wg.Wait()
```

</details>

That's a lot of code right there for a simple work pool! We even had to make it collapsable to avoid disrupting the reading flow. Admittedly, most of the complexity comes from cancellation handling, but you don't want to go around leaking your goroutines.

Now let's see how the same thing is done with JPipe:

```go
ids := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
pipeline := jpipe.New(ctx)
<-jpipe.FromSlice(pipeline, ids).
    ForEach(expensiveIOOperation, jpipe.Concurrent(5))
```

That was easy, wasn't it? We created a pipeline, created a channel from a slice, and ran the expensive operation for each value, with a concurrency factor of 5.

This is only a glimpse of what you can do with JPipe though. We have a lot of pipeline operators to make your concurrent code readable and safe, so go ahead with this guide for more.

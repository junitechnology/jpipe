---
layout: default
title: Self-stopping Operators
nav_order: 4
parent: Usage
---

<h1>Self-stopping Operators</h1>

Some operators stop themselves at some point. Our example pipeline uses `Take`:

```go
pipeline := jpipe.New(ctx)
<-jpipe.FromRange(pipeline, 1, 1000).
    Filter(func(id int) bool { return !idExists(id) }).
    Take(5).
    ForEach(func(id int) { createEntityWithId(id) })
```

The `Take(5)` operator stops itself after it has processed 5 elements. Stopping itself involves two important steps:

- **It closes its output channel:** When `ForEach` realizes its input channel is closed, it will stop itself too.
- **It unsubscribes from the input channel:** When `Filter` realizes it has no downstream operator subscribed, it will stop itself too. This eventually makes `FromRange` stop. At that point, all operators are stopped, so the pipeline itself will be stopped, and `pipeline.Done()`'s channel will be closed.

So the main lesson here is: **self-stopping operators can potentially stop the whole pipeline**. They are particularly useful when the source(`FromRange` here) may produce a lot of data of which only a subset is needed. Some source operators like `FromGenerator` produce a virtually unlimited number of items, and in such cases, only a self-stopping operator or a manual `pipeline.Cancel(error)` can stop the pipeline.
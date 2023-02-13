---
layout: default
title: Lazy Execution
nav_order: 6
parent: Usage
---

<h1>Lazy Execution</h1>

If you create a `Pipeline` with `jpipe.New(ctx)`, operators won't start working until a sink operator is created. Take e.g.:

```go
pipeline := jpipe.New(ctx)
channel := jpipe.FromRange(pipeline, 1, 1000).
    Filter(func(id int) bool { return !idExists(id) }).
    Take(5)
```

After the above snippet, nothing will be happening at all yet. `FromRange`, `Filter` and `Take` are waiting to be started. If you now add a sink operator:

```go
<-channel.ForEach(func(id int) { createEntityWithId(id) })
```

The pipeline will automatically start, and so will all operators.

This lazy behavior allows for the work to start at a well defined time. But also, it allows to "prepare" pipelines and pass them around as argument for someone else to consume them with a sink operator. In such a case, the consumer most probably wants the pipeline to start when the sink operator is added.

<h2>Starting pipelines manually</h2>

The prepared pipeline may contain the sink operator itself. In that case, auto-start would trigger execution before the pipeline consumer receives the pipeline.

```go
pipeline := jpipe.NewPipeline(ctx, jpipe.Config{Context: ctx, StartManually: true})
jpipe.FromRange(pipeline, 1, 1000).
    Filter(func(id int) bool { return !idExists(id) }).
    Take(5).
    ForEach(func(id int) { createEntityWithId(id) })

// scheduling the pipeline for future execution
time.After(5*time.Minute, pipeline.Start())
```

Notice how in the snippet above, if we had used `jpipe.New(ctx)`, the pipeline would have executed immediately, instead of after 5 minutes.

This whole concept of prepared pipelines can be better seen as the pipeline just being a blueprint of the actual processing. The pipeline defines data generation/flow/transformation, but actual execution is a separate step.
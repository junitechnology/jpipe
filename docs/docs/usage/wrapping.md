---
layout: default
title: Wrapping
nav_order: 7
parent: Usage
---

<h1>Wrapping</h1>

We provide a convenience type `item.Item` to wrap pipeline items that can either have a `Value` or an `Error`:

```go
type Item[T any] struct {
    Value T
    Error error
    Ctx   context.Context
}
```

Any `Channel[T]` can be easily converted to a `Channel[Item[T]]` by calling `Wrap`:

```go
wrappedChannel := jpipe.Wrap(channel)
```

`jpipe.Item` is useful for two things: error handling and context passing. Let's look into those.


<h2>Error handling</h2>

Pipeline errors occur when the pipeline gets canceled, either automatically by context cancelation, or manually by calling `pipeline.Cancel(error)`. You may decide to call `pipeline.Cancel(error)` if you find some condition that you deem as a critical error in the pipeline. Let's say you need to do 20 calls to a rate-limited API, and you want to stop the pipeline if one call returns a rate-limit error:

```go
channel := jpipe.FromRange(pipeline, 1, 20).
    ForEach(func(id int) {
        value, err := getValueFromAPI(id)
        if err != nil {
            fmt.Printf("Error calling API %v\n", err)
            if err.Error() == "rate limit exceeded" {
                pipeline.Cancel(err)
            }
            return
        }
        fmt.Printf("Got value %s from the API for id %d\n", value, id)
    })
```

But not all errors are critical, and non-critical errors should probably not cancel the pipeline. Instead, it may be interesting to pass them along the pipeline:

```go
channel := jpipe.FromRange(pipeline, 1, 20)
wrappedChannel := jpipe.Wrap(channel)
<-jpipe.Map(wrappedChannel, func(it item.Item[int]) item.Item[string] {
        value, err := getValueFromAPI(id)
        return item.Item[string]{Value: value, Error: err, Ctx: it.Ctx}
    }).
    Filter(func(item item.Item[string]) bool {
        if item.Error != nil {
            fmt.Printf("Error calling API %v\n", item.Error)
            return false
        }
        return true
    }).
    ForEach(func(item item.Item[string]) { saveValueToDB(item.Value) })
```

<h2>Context passing</h2>

We've already seen that `item.Item` contains a context. It is useful in case you want to enrich that context as items pass along the pipeline. Let's take the same snippet as before, but enrich the context with some data:

```go
channel := jpipe.FromRange(pipeline, 1, 20)
wrappedChannel := jpipe.Wrap(channel)
<-jpipe.Map(wrappedChannel, func(it item.Item[int]) item.Item[string] {
        value, err := getValueFromAPI(id)
        ctx := context.WithValue(item.Context(), "id", id)
        return item.Item[string]{Value: value, Error: err, Ctx: ctx}
    }).
    Filter(func(item item.Item[string]) bool {
        if item.Error != nil {
            fmt.Printf("Error calling API %v\n", item.Error)
            return false
        }
        return true
    }).
    ForEach(func(item item.Item[string]) {
        fmt.Printf("Saving value to DB for id %s", item.Context.Value("id"))
        saveValueToDB(item.Value)
    })
```

We don't encourage passing data throught the context if it's part of the logical flow of the pipeline, but it can be useful to have some values in the context for some cross-cutting concerns like logging and tracing.

<h2>Custom item types</h2>

You may need a more specialized `Item` type than `item.Item`. In that case, don't hesitate to create your own struct, suitable for your specific needs. `item.Item` is not meant to get in your way.
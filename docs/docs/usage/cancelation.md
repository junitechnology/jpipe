---
layout: default
title: Cancellation
nav_order: 1
parent: Usage
---

<h1>Cancellation</h1>

One of the main points of using JPipe is getting cancellation handled automatically. If you understand the importance of cancellation, or don't care much about it, feel free to skip this page. Otherwise, continue reading so you can understand the problem we are solving.

Simply put, using goroutines and channels without proper cancellation may lead to a goroutine leak, and eventually resource exhaustion. This may seem like a remote problem, but it's actually very easy to run into it:

```go
func handleRequest(ctx context.Context) error {
    select {
    case apiResponse := <-callAPI():
        fmt.Printf("API response: %s\n", apiResponse)
        return nil
    case ctx.Done():
        fmt.Println("Context done before getting API response")
        return ctx.Error()
    }
}

func callAPI() <-chan string {
    resultChannel := make(chan string)
    go resultChannel<-doExpensiveAPICall()
}
```

If the context times out before `callAPI` gets its response, `handleRequest` will just return. But now there's no one to listen on `callAPI`'s returned channel, so the goroutine started by it will be forever stuck in the statement `resultChannel<-doExpensiveAPICall()`. Of course, the solution in this case is passing the context to `callAPI` and checking for `ctx.Done()` there.

<h2>Concurrent worker pools</h2>

Since the main use case for JPipe is concurrent loads, let's see how mishandled cancellation can go wrong there. Let's imagine a requet handler now that executes some concurrent work.

```go
func handleRequest(ctx context.Context) error {
    ids := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    select {
    case err := <-doConcurrentWork(ctx, ids, 5):
        return err
    case ctx.Done():
        return ctx.Error()
    }
}
```

The naive implementation goes like this:

```go
func doConcurrentWork(ctx context.Context, ids []int, concurrency int) <-chan error {
    channel := make(chan int)
    var wg sync.WaitGroup
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for value := range channel {
                expensiveIOOperation(id)
            }
        }()
    }

    done := make(chan bool)
    go func() {
        for _, id := range ids {
            channel <- id
        }
        close(channel)

        wg.Wait()
        done<-nil
    }()

    return done
}
```

What would happen if the context times out before `doConcurrentWork` finishes? Well, `handleRequest` will immediately return. As for `doConcurrentWork`, since it does not check `ctx.Done()` internally, the worker pool will just go ahead and execute all 10 operations. The pool goroutines will complete, and the main goroutine that was waiting on `wg.Wait()` will proceed. It will now try to send to the `done` channel, but that channel is not being read by `handleRequest` anymore, cause that function returned. The goroutine is hopelessly blocked now.

Let's listen to `ctx.Done` in `doConcurrentWork` then:

```go
func doConcurrentWork(ctx context.Context, ids []int, concurrency int) <-chan error {
    channel := make(chan int)
    var wg sync.WaitGroup
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for value := range channel {
                expensiveIOOperation(id)
            }
        }()
    }

    done := make(chan bool)
    go func() {
        for _, id := range ids {
            channel <- id
        }
        close(channel)

        wg.Wait()
        select {
        case done<-nil:
        case ctx.Done():
            done<-ctx.Error()
        }
    }()

    return done
}
```

That was easy. The goroutine will no longer block on `done<-nil`, since it will complete when the context is done. This was not a timely cleanup though. We still waited for all 10 operations to complete. Maybe the context expired after 10 seconds, but the whole work took 30 seconds. So we had 5+1 goroutine working for 20 seconds, but they shouldn't, cause the context expiration should cause all work spawned by it to shut down as soon as possible.

Let's improve on that and listen to `ctx.Done()` before sending items to the worker pool:

```go
func doConcurrentWork(ctx context.Context, ids []int, concurrency int) <-chan bool {
    channel := make(chan int)
    var wg sync.WaitGroup
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for value := range channel {
                expensiveIOOperation(id)
            }
        }()
    }

    done := make(chan bool)
    go func() {
        for _, id := range ids {
            select {
            case channel <- id:
            case <-ctx.Done():
                break
            }
        }
        close(channel)

        wg.Wait()
        select {
        case done<-nil:
        case ctx.Done():
            done<-ctx.Error()
        }
    }()

    return done
}
```

That's better. We no longer need work for all 10 operations to be finished for the goroutines to complete. But there's a gotcha. The `select` statement knows no priority, so even if `ctx.Done()` receives, there are equal chances for the next `channel <- id` to be executed instead. So we are again at the same spot: we are initiating new work after the context times out.

There's a simple trick to give priority to `ctx.Done()` though:

```go
func doConcurrentWork(ctx context.Context, ids []int, concurrency int) <-chan bool {
    channel := make(chan int)
    var wg sync.WaitGroup
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for value := range channel {
                expensiveIOOperation(id)
            }
        }()
    }

    done := make(chan bool)
    go func() {
        for _, id := range ids {
            select {
            case <-ctx.Done():
                break
            default:
                select {
                case channel <- id:
                case <-ctx.Done():
                    break
                }
            }
        }
        close(channel)

        wg.Wait()
        select {
        case done<-nil:
        case ctx.Done():
            done<-ctx.Error()
        }
    }()

    return done
}
```

Now, every time an `expensiveIOOperation(id)` completes, we first check `ctx.Done()` and then proceed to the normal select.

We can probably agree that not only is it a lot of boilerplate, it's also a very tricky code to get right. Fortunately, the same result can be easily achieved with JPipe:

```go
func doConcurrentWork(ctx context.Context, ids []int, concurrency int) <-chan bool {
    pipeline := jpipe.New(ctx)
    jpipe.FromSlice(ids).
        ForEach(expensiveIOOperation)

    return pipeline.Error()
}
```
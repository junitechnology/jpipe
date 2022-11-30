---
layout: default
title: Concurrency
nav_order: 4
parent: Usage
---

<h1>Concurrency</h1>

Some operators accept a `jpipe.Concurrent(int)` option to make them run concurrently. They are all operators that take a function as parameter, since that function could benefit from concurrent execution to process all inputs faster. Operators that don't take a function as parameter, like `Take` or `Buffer`, don't support concurrency, because they just pass values from input to output without any CPU or IO bound work involved.

Let's take `Map` as an example:

```go
channel := jpipe.FromRange(pipeline, 1, 10)
<-jpipe.Map(channel, func(i int) int {
        time.Sleep(time.Second)
        return i*10
    }, jpipe.Concurrent(5)).
    ForEach(func(i int) { fmt.Println(i) })
```

We use `time.Sleep` above to simulate a long-running operation. If we hadn't passed `jpipe.Concurrent(5)` to `Map`, the pipeline would take 10 seconds to finish. With `jpipe.Concurrent(5)`, it takes only two seconds, cause the 10 input values are distributed to 5 goroutines that run concurrently.

If your long-running function is CPU-bound, you should probably use `jpipe.Concurrent(runtime.NumCPU())`. Using a concurrency larger than the number of CPUs would be useless, cause the goroutines would be contending for CPU time.

If your function is IO-bound instead, the right amount of concurrency depends on the external system you are communicating with. For SQL databases, e.g., it doesn't make sense to use much more concurrency than the number of CPUs in the database. For APIs, the right concurrency may depend on the API's rate limits. In general, you will have to test and see what's the right amount of concurrency for your use case.

<h2>Ordering</h2>

The snippet above returns output like this:

```
40 30 50 10 20 60 70 90 80 100 
```

A drawback of concurrency is that you lose ordering guarantees, since the work is no longer executed serially. In the example above, the output order is pretty much random, since the duration of the long-running function is always the same. In real-life situations, where different input values may need longer execution times, the output order will be related to every input's execution time. Faster-to-process inputs will be output fist.

There are many situations where you won't really care about maintaining order, but sometimes it may be necessary for downstream operators to process values in the original order. For those cases, we got you covered:

```go
channel := jpipe.FromRange(pipeline, 1, 10)
<-jpipe.Map(channel, func(i int) int {
        time.Sleep(time.Second)
        return i*10
    }, jpipe.Concurrent(5), jpipe.Ordered(0)).
    ForEach(func(i int) { fmt.Println(i) })
```
```
Output:
10 20 30 40 50 60 70 80 90 100 
```

`jpipe.Ordered(orderBufferSize int)` ensures that output values are emitted in the same order of input values. To do this, we keep an internal buffer of output values, and we serialize their output order. In our example, that means that if 30 is the first value to be output, it won't be directly sent to `ForEach`. Instead, it will be on an internal buffer, waiting for 10 and 20 to be emitted.

This internal buffer creates backpressure. If 10 takes a long time to be output, (20, 30, 40, 50) will be waiting for it in the buffer, and no other inputs will be read in the meantime, cause they wouldn't fit in the buffer. To relieve this backpressure, use a larger `orderBufferSize` like `jpipe.Ordered(5)`. Notice that the actual buffer size is `concurrency + orderBufferSize`, cause we need the buffer size to be `concurrency` as a minimum.

This backpressure means that sometimes there will be idle goroutines waiting for a slow input to be processed. Also, the downstream operators may be idle too waiting for values to be output. For this reason, use `jpipe.Ordered` only if you absolutely need the ordering guarantee. Unordered concurrency will always be faster, since inputs are processed by the first available goroutine, and values output immediately.
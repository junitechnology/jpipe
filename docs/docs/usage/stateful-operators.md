---
layout: default
title: Stateful operators
nav_order: 9
parent: Usage
---

<h1>Stateful operators</h1>

Some operators are inherently stateful. `Distinct` e.g. keeps an internal map with all distinct values seen. `Reduce` keeps the state internally and updates it on every value with the reducer function.

Many other operators, like `Filter` and `Map` are inherently stateless. They execute a function to filter/map values and pass them to the output channel, and they keep no internal state at all. That doesn't mean they can't be made stateful though. One can simply provide an external state in the function closure. Let's see a couple of examples.

`Distinct` can easily be implemented by using `Filter` in a stateful manner:

```go
allSeen := map[any]bool{}
jpipe.FromGenerator(pipeline, func(i uint64) int { return rand.Intn(100) }).
    Filter(func(x int) bool{
        if _, seen := allSeen[x]; seen {
			return false
		}
        seen[x] = true
        return true
    })
```

The full distinct logic could be encapsulated in a helper function:

```go
func distinctFilter() func(x int) bool {
    allSeen := map[any]bool{}
    return func(x int) bool {
        if _, seen := allSeen[x]; seen {
			return false
		}
        seen[x] = true
        return true
    }
}
```

`distinctFilter` could be easily reused then:

```go
jpipe.FromGenerator(pipeline, func(i uint64) int { return rand.Intn(100) }).
    Filter(distinctFilter())
```

Let's see an example of stateful `Map` now. Imagine we want to transform a stream of integers into a stream of their moving average. We create a `movingAvg` function like this:

```go
func movingAvg(windowSize int32) func(int) float64 {
    window := make([]int, windowSize)
    sum := int64(0)
    index := 0
    currentWindowSize := 0
    return func(x int) float64 {
        sum -= window[index]
        sum += int64(x)
        window[index] = x

        index = (index + 1) % windowSize
        if currentWindowSize < windowSize {
            currentWindowSize++
        }

        return float64(sum) / currentWindowSize
    }
}
```

Usage is then simply:

```go
channel := jpipe.FromGenerator(pipeline, func(i uint64) int { return rand.Intn(100) })
movingAverageChannel := jpipe.Map(channel, movingAvg(5))
```
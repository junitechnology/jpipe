package jpipe

import "github.com/junitechnology/jpipe/options"

func Concurrent(concurrency int) options.Concurrent {
	return options.Concurrent{Concurrency: concurrency}
}

func Buffered(size int) options.Buffered {
	return options.Buffered{Size: size}
}

func KeepFirst() options.Keep {
	return options.Keep{Strategy: options.KEEP_FIRST}
}

func KeepLast() options.Keep {
	return options.Keep{Strategy: options.KEEP_LAST}
}

func getOptions[I any, O any](options []I, defaultOptions O) O {
	for i := range options {
		opt, ok := any(options[i]).(O)
		if ok {
			return opt
		}
	}

	return defaultOptions
}

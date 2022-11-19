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

func getOptions[I any, O any](opts []I, defaultOpt O) O {
	for i := range opts {
		if opt, ok := any(opts[i]).(O); ok {
			return opt
		}
	}

	return defaultOpt
}

func getNodeOptions[O any](opts []O) []options.NodeOptions {
	nodeOpts := []options.NodeOptions{}
	for _, opt := range opts {
		if nodeOpt, ok := any(opt).(options.NodeOptions); ok {
			nodeOpts = append(nodeOpts, nodeOpt)
		}
	}

	return nodeOpts
}

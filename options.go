package jpipe

import "github.com/junitechnology/jpipe/options"

func Concurrent(concurrency int) options.Concurrent {
	return options.Concurrent{Concurrency: concurrency}
}

func Ordered(orderBufferSize int) options.Ordered {
	return options.Ordered{OrderBufferSize: orderBufferSize}
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

func getOption[I any, O any](opts []I) *O {
	for i := range opts {
		if opt, ok := any(opts[i]).(O); ok {
			return &opt
		}
	}

	return nil
}

func getOptionOrDefault[I any, O any](opts []I, defaultOpt O) O {
	opt := getOption[I, O](opts)
	if opt == nil {
		return defaultOpt
	}

	return *opt
}

func getNodeOptions[O any](opts []O) []options.NodeOption {
	return mapOptions[O, options.NodeOption](opts)
}

func getPooledWorkerOptions[O any](opts []O) []options.PooledWorkerOption {
	return mapOptions[O, options.PooledWorkerOption](opts)
}

func mapOptions[O any, R any](opts []O) []R {
	nodeOpts := []R{}
	for _, opt := range opts {
		if nodeOpt, ok := any(opt).(R); ok {
			nodeOpts = append(nodeOpts, nodeOpt)
		}
	}

	return nodeOpts
}

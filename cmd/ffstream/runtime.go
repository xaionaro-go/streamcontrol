package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xaionaro-go/streamctl/pkg/observability"
)

func initRuntime(
	ctx context.Context,
	flags Flags,
) (context.Context, context.CancelFunc) {
	var closeFuncs []func()

	l := logger.FromCtx(ctx)

	if flags.ListenNetPprof != "" {
		observability.Go(ctx, func() {
			http.Handle(
				"/metrics",
				promhttp.Handler(),
			) // TODO: either split this from pprof argument, or rename the argument (and re-describe it)

			l.Infof("starting to listen for net/pprof requests at '%s'", flags.ListenNetPprof)
			l.Error(http.ListenAndServe(flags.ListenNetPprof, nil))
		})
	}

	ctx, cancelFn := context.WithCancel(ctx)
	return ctx, func() {
		defer belt.Flush(ctx)
		cancelFn()
		for i := len(closeFuncs) - 1; i >= 0; i-- {
			closeFuncs[i]()
		}
	}
}

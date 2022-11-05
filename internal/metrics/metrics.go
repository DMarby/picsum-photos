package metrics

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/DMarby/picsum-photos/internal/handler"
	"github.com/DMarby/picsum-photos/internal/health"
	"github.com/DMarby/picsum-photos/internal/logger"
)

// Serve starts an http server for metrics and healthchecks
func Serve(ctx context.Context, log *logger.Logger, healthChecker *health.Checker, listenAddress string) {
	router := http.NewServeMux()
	router.HandleFunc("/metrics", handler.VarzHandler)
	router.Handle("/health", handler.Health(healthChecker))

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr:    listenAddress,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Infof("shutting down the metrics http server: %s", err)
		}
	}()

	log.Infof("metrics http server listening on %s", listenAddress)

	<-ctx.Done()

	if err := server.Close(); err != nil {
		log.Warnf("error shutting down metrics http server: %s", err)
	}
}

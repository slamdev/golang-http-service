package integration

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

func WaitForShutdown(listeners ...func() error) error {
	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)
	sig := <-gracefulStop
	slog.Info("shutdown signal received", "signal", sig.String())

	wg, _ := errgroup.WithContext(context.TODO())
	for _, listener := range listeners {
		wg.Go(listener)
	}
	return wg.Wait()
}

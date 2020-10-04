package pkg

import (
	"context"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

func WaitForShutdown(listeners ...func() error) error {
	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)
	sig := <-gracefulStop
	zap.L().Info("shutdown signal received", zap.String("signal", sig.String()))

	wg, _ := errgroup.WithContext(context.TODO())
	for _, listener := range listeners {
		wg.Go(listener)
	}
	return wg.Wait()
}

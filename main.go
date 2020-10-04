package main

import (
	"go.uber.org/zap"
	"golang-http-service/internal"
	"golang-http-service/pkg"
	"os"
)

func main() {
	app, err := internal.NewApp()
	if err != nil {
		zap.L().Error("failed to create app", zap.Error(err))
		os.Exit(1)
	}

	go func() {
		if err := app.Start(); err != nil {
			zap.L().Error("failed to start app", zap.Error(err))
			os.Exit(1)
		}
	}()

	if err := pkg.WaitForShutdown(app.Stop); err != nil {
		zap.L().Error("failed to stop app", zap.Error(err))
		os.Exit(1)
	}

	zap.L().Info("app is stopped")
}

package main

import (
	"fmt"
	"go.uber.org/zap"
	"golang-http-service/pkg"
	"golang-http-service/pkg/integration"
	"os"
)

func main() {
	app, err := pkg.NewApp()
	if err != nil {
		if integration.IsLoggerInitialized() {
			zap.L().Error("failed to create app", zap.Error(err))
		} else {
			fmt.Printf("failed to create app: %+v", err)
		}
		os.Exit(1)
	}

	go func() {
		if err := app.Start(); err != nil {
			zap.L().Error("failed to start app", zap.Error(err))
			os.Exit(1)
		}
	}()

	if err := integration.WaitForShutdown(app.Stop); err != nil {
		zap.L().Error("failed to stop app", zap.Error(err))
		os.Exit(1)
	}

	zap.L().Info("app is stopped")
}

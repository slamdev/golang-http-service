package main

import (
	"log/slog"
	"os"

	"golang-http-service/pkg"
	"golang-http-service/pkg/integration"
	_ "golang.org/x/mod/modfile" // transitive dependency that is not recognized by go mod tidy
)

func main() {
	app, err := pkg.NewApp()
	if err != nil {
		slog.Error("failed to create app", "err", err)
		os.Exit(1)
	}

	go func() {
		if err := app.Start(); err != nil {
			slog.Error("failed to start app", "err", err)
			os.Exit(1)
		}
	}()
	slog.Info("app is running")

	if err := integration.WaitForShutdown(app.Stop); err != nil {
		slog.Error("failed to stop app", "err", err)
		os.Exit(1)
	}

	slog.Info("app is stopped")
}

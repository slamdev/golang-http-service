package internal

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang-http-service/api"
	"golang-http-service/pkg"
	"golang.org/x/sync/errgroup"
)

type App interface {
	Start() error
	Stop() error
}

type app struct {
	config         Config
	actuatorServer pkg.ActuatorServer
	httpServer     pkg.HttpServer
}

func NewApp() (App, error) {
	app := app{}
	if err := pkg.PopulateConfig(&app.config); err != nil {
		return nil, fmt.Errorf("failed to populate config; %w", err)
	}

	if err := pkg.ConfigureLogger(app.config.Logger.Production); err != nil {
		return nil, fmt.Errorf("failed to configure logger; %w", err)
	}

	zap.S().Infow("starting app", "config", app.config)

	app.actuatorServer = pkg.NewActuatorServer(app.config.Actuator.Port)

	app.httpServer = pkg.NewHttpServer(app.config.Http.Port, func(echo *echo.Echo) {
		api.RegisterHandlers(echo, NewController())
	})

	return &app, nil
}

func (a *app) Start() error {
	done := make(chan error, 2)
	go func() {
		done <- a.actuatorServer.Start()
	}()
	go func() {
		done <- a.httpServer.Start()
	}()
	for i := 0; i < cap(done); i++ {
		if err := <-done; err != nil {
			return err
		}
	}
	return nil
}

func (a *app) Stop() error {
	ctx := context.TODO()
	wg, _ := errgroup.WithContext(ctx)
	wg.Go(func() error { return a.actuatorServer.Stop(ctx) })
	wg.Go(func() error { return a.httpServer.Stop(ctx) })
	return wg.Wait()
}

type Config struct {
	Http struct {
		Port int32
	}
	Actuator struct {
		Port int32
	}
	Logger struct {
		Production bool
	}
}

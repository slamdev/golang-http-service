package pkg

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang-http-service/api"
	"golang-http-service/pkg/business/boundary"
	"golang-http-service/pkg/business/control"
	"golang-http-service/pkg/integration"
	"golang.org/x/sync/errgroup"
	"net/http"
)

type App interface {
	Start() error
	Stop() error
}

type app struct {
	config         Config
	actuatorServer integration.ActuatorServer
	httpServer     integration.HttpServer
}

func NewApp() (App, error) {
	app := app{}
	if err := integration.PopulateConfig(&app.config); err != nil {
		return nil, fmt.Errorf("failed to populate config; %w", err)
	}

	if err := integration.ConfigureLogger(app.config.Logger.Production); err != nil {
		return nil, fmt.Errorf("failed to configure logger; %w", err)
	}

	zap.S().Infow("starting app", "config", app.config)

	app.actuatorServer = integration.NewActuatorServer(app.config.Actuator.Port)

	userRepo := control.NewUserRepo()
	controller := boundary.NewController(userRepo)

	app.httpServer = integration.NewHttpServer(app.config.Http.Port, func(echo *echo.Echo) {
		handler := api.NewStrictHandler(controller, []api.StrictMiddlewareFunc{validationMiddleware})
		api.RegisterHandlersWithBaseURL(echo, handler, app.config.BaseUrl)
	})

	return &app, nil
}

func validationMiddleware(f api.StrictHandlerFunc, _ string) api.StrictHandlerFunc {
	return func(ctx echo.Context, i interface{}) (interface{}, error) {
		if err := ctx.Validate(i); err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err)
		}
		return f(ctx, i)
	}
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
	BaseUrl string `yaml:"baseUrl"`
}

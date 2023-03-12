package pkg

import (
	"context"
	"fmt"
	"github.com/go-logr/zapr"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
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
	traceProvider  *trace.TracerProvider
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

	if tp, err := initTracer(true); err != nil {
		return nil, fmt.Errorf("failed to init tracer; %w", err)
	} else {
		app.traceProvider = tp
	}

	app.httpServer = integration.NewHttpServer(app.config.Http.Port, func(echo *echo.Echo) {
		handler := api.NewStrictHandler(controller, []api.StrictMiddlewareFunc{validationMiddleware})
		api.RegisterHandlersWithBaseURL(echo, handler, app.config.BaseUrl)
	})

	return &app, nil
}

func initTracer(useStdoutExporter bool) (*trace.TracerProvider, error) {
	var exporter trace.SpanExporter
	if useStdoutExporter {
		writer := &zapio.Writer{Log: zap.L(), Level: zap.DebugLevel}
		var err error
		if exporter, err = stdouttrace.New(stdouttrace.WithWriter(writer)); err != nil {
			return nil, err
		}
	} else {
		var err error
		if exporter, err = otlptracegrpc.New(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
	}
	staticRes, err := resource.Merge(resource.Default(), resource.NewSchemaless(semconv.ServiceName("app")))
	if err != nil {
		return nil, fmt.Errorf("to merge static resources: %w", err)
	}
	otelRes, err := resource.Merge(staticRes, resource.Environment())
	if err != nil {
		return nil, fmt.Errorf("to merge env resources: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(otelRes),
	)

	otel.SetTracerProvider(tp)
	otel.SetLogger(zapr.NewLogger(zap.L()))
	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())
	return tp, nil
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
	wg.Go(func() error { return a.traceProvider.Shutdown(ctx) })
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

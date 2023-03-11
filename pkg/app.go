package pkg

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"golang-http-service/api"
	"golang-http-service/pkg/business/boundary"
	"golang-http-service/pkg/business/control"
	"golang-http-service/pkg/integration"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"time"
)

type App interface {
	Start() error
	Stop() error
}

type app struct {
	config         Config
	actuatorServer integration.ActuatorServer
	httpServer     integration.HttpServer
	traceProvider  *sdktrace.TracerProvider
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

	if tp, err := initTracer(); err != nil {
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

func initTracer() (*sdktrace.TracerProvider, error) {
	//writer := &zapio.Writer{Log: zap.L(), Level: zap.DebugLevel}
	//exporter, err := stdout.New(stdout.WithWriter(writer))
	//if err != nil {
	//	return nil, err
	//}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "localhost:4317",
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
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

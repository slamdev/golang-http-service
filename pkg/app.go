package pkg

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexliesenfeld/health"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang-http-service/pkg/business/boundary"
	"golang-http-service/pkg/business/control"
	"golang-http-service/pkg/integration"
)

type App interface {
	Start() error
	Stop() error
}

type app struct {
	config         integration.Config
	actuatorServer integration.HttpServer
	apiServer      integration.HttpServer
	traceProvider  *trace.TracerProvider
	metricProvider *metric.MeterProvider
}

func NewApp() (App, error) {
	ctx := context.TODO()
	app := app{}
	if err := integration.PopulateConfig(&app.config); err != nil {
		return nil, fmt.Errorf("failed to populate config; %w", err)
	}

	telemetryResource, err := integration.CreateTelemetryResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry resource; %w", err)
	}

	if err := integration.ConfigureLogProvider(telemetryResource, app.config.Telemetry.Logs.Level, app.config.Telemetry.Logs.Format); err != nil {
		return nil, fmt.Errorf("failed to init log provider; %w", err)
	}

	if tp, err := integration.ConfigureTraceProvider(ctx, telemetryResource, app.config.Telemetry.Traces.Output); err != nil {
		return nil, fmt.Errorf("failed to init tracer; %w", err)
	} else {
		app.traceProvider = tp
	}

	if mp, err := integration.ConfigureMetricProvider(ctx, telemetryResource, app.config.Telemetry.Metrics.Output); err != nil {
		return nil, fmt.Errorf("failed to init metric provider; %w", err)
	} else {
		app.metricProvider = mp
	}

	_, err = integration.CreatePetStoreAPIClient(app.config.Petstore.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to create harbor client; %w", err)
	}

	userRepo := control.NewUserRepo()
	controller := boundary.NewController(userRepo)

	roleDefs := make(map[string]string)
	for _, role := range app.config.Auth.Roles {
		roleDefs[role.Name] = role.Audience
	}

	apiHandler, err := integration.APIHandler(app.config.BaseUrl, controller, app.config.Auth.Enabled, app.config.Auth.JwkSetUri, app.config.Auth.AllowedIssuers, roleDefs)
	if err != nil {
		return nil, fmt.Errorf("failed to create api handler; %w", err)
	}
	app.apiServer = integration.NewHttpServer(app.config.Http.Port, apiHandler)

	exampleCheck := health.Check{
		Name:  "db",
		Check: func(ctx context.Context) error { return nil },
	}

	app.actuatorServer = integration.NewHttpServer(app.config.Actuator.Port, integration.TelemetryHandler(exampleCheck))
	return &app, nil
}

func (a *app) Start() error {
	starters := []func() error{
		a.actuatorServer.Start,
		a.apiServer.Start,
	}
	done := make(chan error, len(starters))
	for i := range starters {
		go func() { done <- starters[i]() }()
	}
	for i := 0; i < cap(done); i++ {
		if err := <-done; err != nil {
			return err
		}
	}
	return nil
}

func (a *app) Stop() error {
	ctx := context.TODO()
	return errors.Join(
		a.actuatorServer.Stop(ctx),
		a.apiServer.Stop(ctx),
		a.traceProvider.Shutdown(ctx),
		a.metricProvider.Shutdown(ctx),
	)
}

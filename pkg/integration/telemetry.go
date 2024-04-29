package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alexliesenfeld/health"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/felixge/httpsnoop"
	"github.com/go-logr/logr"
	"github.com/lmittmann/tint"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	slogotel "github.com/remychantenay/slog-otel"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

func CreateTelemetryResource(ctx context.Context) (*resource.Resource, error) {
	res, err := resource.New(ctx,
		resource.WithContainer(),
		//resource.WithHost(),
		//resource.WithProcess(),
		//resource.WithTelemetrySDK(),
		resource.WithSchemaURL("https://opentelemetry.io/schemas/1.7.0"),
		resource.WithContainerID(),
		//resource.WithOS(),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}
	merged, err := resource.Merge(res, resource.NewSchemaless(semconv.ServiceName("app")))
	if err != nil {
		return nil, fmt.Errorf("failed to merge static resources: %w", err)
	}

	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		slog.Error("otel error", "err", err)
	}))
	return merged, nil
}

func ConfigureTraceProvider(ctx context.Context, res *resource.Resource, output string) (*trace.TracerProvider, error) {
	var exporter trace.SpanExporter
	if output == "remote" {
		var err error
		if exporter, err = otlptracegrpc.New(ctx); err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
	} else {
		var writer io.Writer
		if output == "noop" {
			writer = NoopWriter{}
		} else {
			writer = &SlogWriter{Log: slog.Default(), Level: slog.LevelDebug}
		}
		var err error
		if exporter, err = stdouttrace.New(stdouttrace.WithWriter(writer)); err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

func ConfigureMetricProvider(_ context.Context, res *resource.Resource, output string) (*metric.MeterProvider, error) {
	var reader metric.Reader
	if output == "remote" {
		// recreate default registry to remove built-in collectors
		// they are covered by otel
		reg := prometheus.NewRegistry()
		prometheus.DefaultRegisterer = reg
		prometheus.DefaultGatherer = reg
		var err error
		if reader, err = promexporter.New(); err != nil {
			return nil, fmt.Errorf("failed to create metric exporter: %w", err)
		}
	} else {
		var writer io.Writer
		if output == "noop" {
			writer = NoopWriter{}
		} else {
			writer = &SlogWriter{Log: slog.Default(), Level: slog.LevelDebug}
		}
		exporter, err := stdoutmetric.New(stdoutmetric.WithWriter(writer))
		if err != nil {
			return nil, fmt.Errorf("failed to create metric exporter: %w", err)
		}
		reader = metric.NewPeriodicReader(exporter)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	if err := host.Start(); err != nil {
		return nil, fmt.Errorf("failed to start host observer: %w", err)
	}

	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(5 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to start runtime observer: %w", err)
	}

	return mp, nil
}

// ConfigureLogProvider replace with OTEL log bridge when it's GA
func ConfigureLogProvider(_ *resource.Resource, level string, format string) error {
	lvl := slog.LevelDebug
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		fmt.Printf("failed to parse log level: %v, fallback to DEBUG", err)
	}
	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	} else {
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
		})
	}

	l := slog.New(slogotel.OtelHandler{Next: handler})

	slog.SetDefault(l)
	otel.SetLogger(logr.FromSlogHandler(l.Handler()))
	return nil
}

func TelemetryHandler(checks ...health.Check) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleHTTPNotFound)
	mux.Handle("/metrics", PrometheusHandler())
	mux.Handle("/health", HealthCheckHandler(checks...))
	h := RecoverMiddleware(mux)
	return h
}

func PrometheusHandler() http.Handler {
	writer := &SlogWriter{Log: slog.Default(), Level: slog.LevelError}
	return promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		ErrorLog:          log.New(writer, "", 0),
		ErrorHandling:     promhttp.HTTPErrorOnError,
		Timeout:           5 * time.Second,
		EnableOpenMetrics: true,
		ProcessStartTime:  time.Now(),
	})
}

func HealthCheckHandler(checks ...health.Check) http.Handler {
	var checkOptions []health.CheckerOption
	checkOptions = append(checkOptions, health.WithTimeout(3*time.Second))
	checkOptions = append(checkOptions, health.WithStatusListener(healthStatusListener))
	for _, check := range checks {
		checkOptions = append(checkOptions, health.WithPeriodicCheck(3*time.Second, 1*time.Second, check))
	}
	healthChecker := health.NewChecker(checkOptions...)
	return health.NewHandler(healthChecker)
}

func healthStatusListener(ctx context.Context, state health.CheckerState) {
	var attrs []any
	attrs = []any{slog.String("status", string(state.Status))}
	for name, checkState := range state.CheckState {
		cha := []any{
			slog.String("status", string(checkState.Status)),
			slog.Time("lastCheckedAt", checkState.LastCheckedAt),
			slog.Time("lastSuccessAt", checkState.LastSuccessAt),
			slog.Time("firstCheckStartedAt", checkState.FirstCheckStartedAt),
		}
		if checkState.Result != nil {
			cha = append(cha, slog.Any("error", checkState.Result))
		}
		if !checkState.LastFailureAt.IsZero() {
			cha = append(cha, slog.Time("lastFailureAt", checkState.LastFailureAt))
		}
		if checkState.ContiguousFails > 0 {
			cha = append(cha, slog.Uint64("contiguousFails", uint64(checkState.ContiguousFails)))
		}
		g := slog.Group(name, cha...)
		attrs = append(attrs, g)
	}
	slog.InfoContext(ctx, "health status changed", attrs...)
}

func logHTTPRequest(r *http.Request, m httpsnoop.Metrics) {
	bytesIn, _ := strconv.Atoi(r.Header.Get("Content-Length"))
	attrs := []any{
		slog.String("host", r.Host),
		slog.String("uri", r.RequestURI),
		slog.String("method", r.Method),
		slog.String("referer", r.Referer()),
		slog.Int("status", m.Code),
		slog.Int("bytesIn", bytesIn),
		slog.Int64("bytesOut", m.Written),
		slog.Duration("latency", m.Duration),
	}

	span := oteltrace.SpanFromContext(r.Context())
	if readSpan, ok := span.(trace.ReadOnlySpan); ok {
		for _, event := range readSpan.Events() {
			if event.Name == semconv.ExceptionEventName {
				var errAttrs []any
				for _, a := range event.Attributes {
					errAttrs = append(errAttrs, slog.String(string(a.Key), a.Value.AsString()))
				}
				attrs = append(attrs, slog.Group("err", errAttrs...))
				break
			}
		}
	}

	slog.InfoContext(r.Context(), "access", attrs...)
}

func labelRequest(ctx context.Context, requestID string) {
	var spanAttrs []attribute.KeyValue

	claims, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if ok {
		userIDAttr := semconv.EnduserIDKey.String(claims.RegisteredClaims.Subject)
		spanAttrs = append(spanAttrs, userIDAttr)
		audienceAttr := semconv.EnduserScopeKey.StringSlice(claims.RegisteredClaims.Audience)
		spanAttrs = append(spanAttrs, audienceAttr)
	}

	roles, ok := ctx.Value(AuthRoleKey{}).([]string)
	if ok {
		rolesAttr := semconv.EnduserRoleKey.StringSlice(roles)
		spanAttrs = append(spanAttrs, rolesAttr)
	}

	// copy from
	// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/instrumentation/net/http/otelhttp/v0.49.0/instrumentation/net/http/otelhttp/handler.go#L273-L279
	operationAttr := semconv.HTTPRouteKey.String(requestID)
	spanAttrs = append(spanAttrs, operationAttr)

	span := oteltrace.SpanFromContext(ctx)
	span.SetAttributes(spanAttrs...)

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	labeler.Add(operationAttr)
}

// SlogWriter based on https://github.com/uber-go/zap/blob/v1.27.0/zapio/writer.go
type SlogWriter struct {
	Log   *slog.Logger
	Level slog.Level
	buff  bytes.Buffer
}

func (w *SlogWriter) Write(bs []byte) (n int, err error) {
	if !w.Log.Enabled(context.Background(), w.Level) {
		return len(bs), nil
	}
	n = len(bs)
	for len(bs) > 0 {
		bs = w.writeLine(bs)
	}
	return n, nil
}

func (w *SlogWriter) writeLine(line []byte) (remaining []byte) {
	idx := bytes.IndexByte(line, '\n')
	if idx < 0 {
		w.buff.Write(line)
		return nil
	}
	line, remaining = line[:idx], line[idx+1:]
	if w.buff.Len() == 0 {
		w.log(line)
		return
	}
	w.buff.Write(line)
	w.flush(true)
	return remaining
}

func (w *SlogWriter) Close() error {
	return w.Sync()
}

func (w *SlogWriter) Sync() error {
	w.flush(false)
	return nil
}

func (w *SlogWriter) flush(allowEmpty bool) {
	if allowEmpty || w.buff.Len() > 0 {
		w.log(w.buff.Bytes())
	}
	w.buff.Reset()
}

func (w *SlogWriter) log(b []byte) {
	w.Log.Log(context.Background(), w.Level, string(b))
}

type NoopWriter struct {
}

func (w NoopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

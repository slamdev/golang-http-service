package integration

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

type ActuatorServer interface {
	Start() error
	Stop(ctx context.Context) error
}

type actuatorServer struct {
	e    *echo.Echo
	port int32
}

func NewActuatorServer(port int32) ActuatorServer {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetHeader("${prefix}")
	e.Logger.SetOutput(zap.NewStdLog(zap.L()).Writer())
	e.Use(middleware.Recover())
	e.GET("/health", handleHealthRequest)
	e.GET("/metrics", handleMetricsRequest)
	return &actuatorServer{e: e, port: port}
}

func (s *actuatorServer) Start() error {
	if err := s.e.Start(fmt.Sprintf(":%d", s.port)); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *actuatorServer) Stop(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}

func handleMetricsRequest(c echo.Context) error {
	h := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{})
	h.ServeHTTP(c.Response(), c.Request())
	return nil
}

func handleHealthRequest(c echo.Context) error {
	r := make(map[string]string)
	r["status"] = "SERVING"
	return c.JSON(http.StatusOK, r)
}

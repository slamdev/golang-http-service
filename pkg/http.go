package pkg

import (
	"context"
	"fmt"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type HttpServer interface {
	Start() error
	Stop(ctx context.Context) error
}

type httpServer struct {
	e    *echo.Echo
	port int32
}

func NewHttpServer(port int32, customizers ...func(echo *echo.Echo)) HttpServer {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetHeader("${prefix}")
	e.Logger.SetOutput(zap.NewStdLog(zap.L()).Writer())
	e.Use(loggerMiddleware())
	e.Use(middleware.Recover())

	p := prometheus.NewPrometheus("http", nil)
	e.Use(p.HandlerFunc)

	for _, c := range customizers {
		c(e)
	}
	return &httpServer{e: e, port: port}
}

func (h *httpServer) Start() (err error) {
	if err = h.e.Start(fmt.Sprintf(":%d", h.port)); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (h *httpServer) Stop(ctx context.Context) error {
	return h.e.Shutdown(ctx)
}

func loggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()
			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)

				httpErr, ok := err.(*echo.HTTPError)
				if ok {
					if httpErr.Internal != nil {
						err = httpErr.Internal
					} else {
						err = fmt.Errorf("%v", httpErr.Message)
					}
				}
			}
			stop := time.Now()

			bytesIn, _ := strconv.Atoi(req.Header.Get(echo.HeaderContentLength))

			zap.L().Info("access",
				zap.String("remote-ip", c.RealIP()),
				zap.String("host", req.Host),
				zap.String("uri", req.RequestURI),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.String("referer", req.Referer()),
				zap.Int("status", res.Status),
				zap.Error(err),
				zap.Duration("latency", stop.Sub(start)),
				zap.Int("bytes-in", bytesIn),
				zap.Int64("bytes-out", res.Size),
			)
			return
		}
	}
}

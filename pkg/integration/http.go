package integration

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/meysamhadeli/problem-details"
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

type CustomValidator struct {
	validator *validator.Validate
}

func NewHttpServer(port int32, customizers ...func(echo *echo.Echo)) HttpServer {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetHeader("${prefix}")
	e.Logger.SetOutput(zap.NewStdLog(zap.L()).Writer())
	e.Use(loggerMiddleware())
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = handleEchoError
	e.Validator = &CustomValidator{validator: validator.New()}

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

func handleEchoError(err error, c echo.Context) {
	if !c.Response().Committed {
		if _, ok := err.(*echo.HTTPError); ok {
			// problem-details library expects echo error Message to be `error` type
			// but it's not always the case, so we set Message to error in case its not
			if _, ok := err.(*echo.HTTPError).Message.(error); !ok {
				err.(*echo.HTTPError).Message = errors.New(err.(*echo.HTTPError).Message.(string))
			}
		}
		if _, err := problem.ResolveProblemDetails(c.Response(), c.Request(), err); err != nil {
			zap.Error(err)
		}
	}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

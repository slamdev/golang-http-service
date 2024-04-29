package integration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang-http-service/api"
)

type HttpServer interface {
	Start() error
	Stop(ctx context.Context) error
}

type httpServer struct {
	srv *http.Server
}

func NewHttpServer(port int32, handler http.Handler) HttpServer {
	srv := http.Server{
		Addr: fmt.Sprintf(":%d", port), Handler: handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return &httpServer{srv: &srv}
}

func (h *httpServer) Start() (err error) {
	if err = h.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (h *httpServer) Stop(ctx context.Context) error {
	return h.srv.Shutdown(ctx)
}

func HandleHTTPBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusBadRequest
	p := createAndRecordProblemDetail(r.Context(), status, err)
	writeProblem(w, r, p)
}

func HandleHTTPNotFound(w http.ResponseWriter, r *http.Request) {
	status := http.StatusNotFound
	p := createAndRecordProblemDetail(r.Context(), status, nil)
	writeProblem(w, r, p)
}

func HandleHTTPUnauthorized(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusUnauthorized
	p := createAndRecordProblemDetail(r.Context(), status, err)
	writeProblem(w, r, p)
}

func HandleHTTPServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "unexpected error occurred", "err", err)

	status := http.StatusInternalServerError
	p := createAndRecordProblemDetail(r.Context(), status, err)
	writeProblem(w, r, p)
}

func writeProblem(w http.ResponseWriter, r *http.Request, p api.ProblemDetail) {
	span := trace.SpanFromContext(r.Context())
	span.RecordError(p.Detail)
	span.SetStatus(codes.Error, p.Title)

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	if err := json.NewEncoder(w).Encode(p); err != nil {
		slog.ErrorContext(r.Context(), "failed to write problem to response", "err", err)
	}
}

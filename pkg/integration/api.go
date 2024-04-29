package integration

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang-http-service/api"
)

func APIHandler(baseURL string, apiController api.StrictServerInterface, enableAuth bool, jwkSetURI string, allowedIssuers []string, roleDefs map[string]string) (http.Handler, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("failed to get embedded swagger spec; %w", err)
	}
	openapiValidationMiddleware := OpenapiValidationMiddleware(swagger)

	middlewares := []api.MiddlewareFunc{openapiValidationMiddleware}
	if enableAuth {
		jwtMiddleware, err := JWTAuthMiddleware(jwkSetURI, allowedIssuers)
		if err != nil {
			return nil, fmt.Errorf("failed to create jwt middleware; %w", err)
		}
		middlewares = append(middlewares, jwtMiddleware)
		middlewares = append(middlewares, AuthRolesMiddleware(roleDefs))
	}

	strictHandler := api.NewStrictHandlerWithOptions(apiController,
		[]api.StrictMiddlewareFunc{TelemetryStrictMiddleware},
		api.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  HandleHTTPBadRequest,
			ResponseErrorHandlerFunc: HandleHTTPServerError,
		},
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleHTTPNotFound)
	h := api.HandlerWithOptions(strictHandler, api.StdHTTPServerOptions{
		BaseURL:          baseURL,
		BaseRouter:       mux,
		Middlewares:      middlewares,
		ErrorHandlerFunc: HandleHTTPBadRequest,
	})
	h = RequestURIMiddleware(RecoverMiddleware(TelemetryGlobalMiddleware(AccessLogsMiddleware(h))))
	return h, nil
}

func NotFoundError(ctx context.Context, err error) api.NotFoundApplicationProblemPlusJSONResponse {
	p := createAndRecordProblemDetail(ctx, http.StatusNotFound, err)
	return api.NotFoundApplicationProblemPlusJSONResponse(p)
}

func BadRequestError(ctx context.Context, err error) api.BadRequestApplicationProblemPlusJSONResponse {
	p := createAndRecordProblemDetail(ctx, http.StatusBadRequest, err)
	return api.BadRequestApplicationProblemPlusJSONResponse(p)
}

func createAndRecordProblemDetail(ctx context.Context, status int, err error) api.ProblemDetail {
	title := http.StatusText(status)
	span := trace.SpanFromContext(ctx)
	var traceID string
	if span.SpanContext().HasTraceID() {
		traceID = span.SpanContext().TraceID().String()
	}
	if err != nil {
		span.RecordError(err)
	}
	span.SetStatus(codes.Error, title)
	requestURI, _ := ctx.Value(RequestURIKey{}).(string)
	return api.ProblemDetail{
		Detail:   err,
		Instance: requestURI,
		Status:   status,
		Title:    title,
		TraceId:  traceID,
		Type:     "about:blank",
	}
}

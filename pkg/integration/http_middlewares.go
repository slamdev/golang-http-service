package integration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/felixge/httpsnoop"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func JWTAuthMiddleware(JwkSetUri string, issuers []string) (func(http.Handler) http.Handler, error) {
	jwkSetURL, err := url.Parse(JwkSetUri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JwkSetUri URL: %w", err)
	}
	provider := jwks.NewCachingProvider(jwkSetURL, 5*time.Minute, jwks.WithCustomJWKSURI(jwkSetURL))
	customClaimsFunc := func() validator.CustomClaims { return &JWTCustomClaims{} }
	validateTokenFunc := func(ctx context.Context, tokenString string) (interface{}, error) {
		return validateToken(ctx, tokenString, issuers, provider.KeyFunc, customClaimsFunc)
	}
	mdl := jwtmiddleware.New(validateTokenFunc, jwtmiddleware.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
		switch {
		case errors.Is(err, jwtmiddleware.ErrJWTMissing):
			HandleHTTPBadRequest(w, r, err)
		case errors.Is(err, jwtmiddleware.ErrJWTInvalid):
			HandleHTTPUnauthorized(w, r, err)
		default:
			HandleHTTPServerError(w, r, err)
		}
	}))
	return func(handler http.Handler) http.Handler { return mdl.CheckJWT(handler) }, nil
}

type AuthRoleKey struct{}

func AuthRolesMiddleware(roleDefs map[string]string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var roles []string
			if claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims); ok {
				if customClaims, ok := claims.CustomClaims.(*JWTCustomClaims); ok {
					for _, claimRole := range customClaims.Roles {
						if audiences, ok := roleDefs[claimRole]; ok && slices.Contains(claims.RegisteredClaims.Audience, audiences) {
							roles = append(roles, claimRole)
						}
					}
				}
			}
			r = r.WithContext(context.WithValue(r.Context(), AuthRoleKey{}, roles))
			next.ServeHTTP(w, r)
		})
	}
}

func OpenapiValidationMiddleware(swagger *openapi3.T) func(next http.Handler) http.Handler {
	options := &nethttpmiddleware.Options{
		SilenceServersWarning: true,
		ErrorHandlerWithOpts: func(w http.ResponseWriter, message string, statusCode int, opts nethttpmiddleware.ErrorHandlerOpts) {
			switch statusCode {
			case http.StatusNotFound:
				HandleHTTPNotFound(w, opts.Request)
			case http.StatusBadRequest:
				HandleHTTPBadRequest(w, opts.Request, errors.New(message))
			default:
				HandleHTTPServerError(w, opts.Request, errors.New(message))
			}
		},
		Options: openapi3filter.Options{
			ExcludeReadOnlyValidations: true,
			AuthenticationFunc:         openapi3filter.NoopAuthenticationFunc,
		},
	}
	return nethttpmiddleware.OapiRequestValidatorWithOptions(swagger, options)
}

func TelemetryStrictMiddleware(f nethttp.StrictHTTPHandlerFunc, operationID string) nethttp.StrictHTTPHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		labelRequest(ctx, operationID)
		return f(ctx, w, r, request)
	}
}

func TelemetryGlobalMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "server", otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents))
}

func AccessLogsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(next, w, r)
		logHTTPRequest(r, m)
	})
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				err := fmt.Errorf("%+v", err)
				HandleHTTPServerError(w, r, err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type RequestURIKey struct{}

func RequestURIMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestURIKey{}, r.RequestURI)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

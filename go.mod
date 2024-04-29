module golang-http-service

go 1.22

// waiting for https://github.com/deepmap/oapi-codegen/commit/dd082985a9b6e8f68472987f6c2c60fea5e59871
// to be released
replace github.com/deepmap/oapi-codegen/v2 => github.com/deepmap/oapi-codegen/v2 v2.1.1-0.20240422103956-472f1cad6201

// oapi-codegen depends on a particular version of kin-openapi
// https://github.com/deepmap/oapi-codegen/blob/master/go.mod#L6
replace github.com/getkin/kin-openapi => github.com/getkin/kin-openapi v0.123.0

// using fork until https://github.com/oapi-codegen/nethttp-middleware/pull/15 is merged
replace github.com/oapi-codegen/nethttp-middleware => github.com/mikeschinkel/nethttp-middleware v0.0.0-20240425122735-247404ba1c72

require (
	github.com/alexliesenfeld/health v0.8.0
	github.com/auth0/go-jwt-middleware/v2 v2.2.1
	github.com/deepmap/oapi-codegen/v2 v2.1.1
	github.com/felixge/httpsnoop v1.0.4
	github.com/getkin/kin-openapi v0.124.0
	github.com/go-faker/faker/v4 v4.4.1
	github.com/go-logr/logr v1.4.1
	github.com/golang/mock v1.6.0
	github.com/lmittmann/tint v1.0.4
	github.com/oapi-codegen/nethttp-middleware v1.0.1
	github.com/oapi-codegen/runtime v1.1.1
	github.com/prometheus/client_golang v1.19.0
	github.com/remychantenay/slog-otel v1.3.0
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/contrib/instrumentation/host v0.51.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.51.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.51.0
	go.opentelemetry.io/contrib/propagators/autoprop v0.51.0
	go.opentelemetry.io/otel v1.26.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.26.0
	go.opentelemetry.io/otel/exporters/prometheus v0.48.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.26.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.26.0
	go.opentelemetry.io/otel/sdk v1.26.0
	go.opentelemetry.io/otel/sdk/metric v1.26.0
	go.opentelemetry.io/otel/trace v1.26.0
	go.uber.org/config v1.4.0
	golang.org/x/mod v0.17.0
	golang.org/x/sync v0.7.0
	gopkg.in/go-jose/go-jose.v2 v2.6.3
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1 // indirect
	github.com/invopop/yaml v0.3.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20240408141607-282e7b5d6b74 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.53.0 // indirect
	github.com/prometheus/procfs v0.14.0 // indirect
	github.com/shirou/gopsutil/v3 v3.24.3 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.8.0 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.26.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.26.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.26.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.26.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.26.0 // indirect
	go.opentelemetry.io/otel/metric v1.26.0 // indirect
	go.opentelemetry.io/proto/otlp v1.2.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.20.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240415180920-8c6c420018be // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240415180920-8c6c420018be // indirect
	google.golang.org/grpc v1.63.2 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

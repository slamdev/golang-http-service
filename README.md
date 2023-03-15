# golang-http-service

TODO:

- adjust readme

Template for creating HTTP service

## Components

### HTTP API

API is defined in [OpenAPI 3](https://swagger.io/specification/) format in the [openapi.yaml](api/openapi.yaml) file.
DTOs and service interface code is generated using [oapi-codegen](https://github.com/deepmap/oapi-codegen).

### HTTP server

[Echo](https://echo.labstack.com/) framework is used to manage routes. API generator has a nice integration with the
framework also

### Configuration

Application configuration is defined [application.yaml](configs/application.yaml) file. There is **profiles** system
that
allows to use and merge multiple configuration files together. It is controlled with **ACTIVE_PROFILES** environment
variable. E.g.: `ACTIVE_PROFILES=cloud,dev` will merge these files together:

1. application.yaml
2. application-cloud.yaml
3. application-dev.yaml

[Uber config](https://github.com/uber-go/config) library is used to parse and merge config files. It also supports
environment variables.

Configuration files are embedded into the resulting binary with [pkger](https://github.com/markbates/pkger) tool.

Additionally you can add more configuration files from filesystem by defining `APP_CONFIG_ADDITIONAL_LOCATION` env variable. 
In this case the app will recursively search for `application.yaml` files in that location. 

### Monitoring

[Zap](https://github.com/uber-go/zap) is used to control logs. Logs are outputted in plain text format when the
application is running locally and in json format when the **cloud** profile is used.

Application exposes [Prometheus](https://prometheus.io/) metrics at **/metrics** endpoint. List of metrics includes
basic http stats collected via [Echo Prometheus](https://github.com/labstack/echo-contrib/tree/master/prometheus)
library.

Application exposes health endpoint at **/health**.

Both prometheus and health endpoints are served on a separate port to make sure it is not exposed to outside world.

### Testing

For reach assertions the [testify](https://github.com/stretchr/testify) library is used.

### Linting

[Spectral](https://github.com/stoplightio/spectral) is used to lint OpenAPI files and
[golangci-lint](https://github.com/golangci/golangci-lint) for Go files.

### Building

All the build process is describe it the (Makefile)[Makefile]. Run `make build` to test and build the binary.

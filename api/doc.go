package api

import _ "github.com/deepmap/oapi-codegen/pkg/runtime"
import _ "github.com/getkin/kin-openapi/openapi3"

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=gen-config-server.yaml openapi.yaml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config=gen-config-client.yaml openapi.yaml

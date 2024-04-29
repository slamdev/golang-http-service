//go:build tools
// +build tools

package petstore

import (
	_ "github.com/deepmap/oapi-codegen/v2/pkg/codegen"
	_ "github.com/getkin/kin-openapi/openapi3"
)

//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=gen-config-client.yaml openapi.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config=gen-config-models.yaml openapi.yaml

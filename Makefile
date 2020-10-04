pkger:
ifeq (, $(shell which pkger))
	go install github.com/markbates/pkger/cmd/pkger
endif
	pkger

openapi:
ifeq (, $(shell which oapi-codegen))
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen
endif
	oapi-codegen -package api -generate "types,server" api/openapi.yaml > api/api.gen.go

generate: openapi pkger

spectral:
ifeq (, $(shell which spectral))
	curl -L https://raw.githack.com/stoplightio/spectral/master/scripts/install.sh | sh
endif
	spectral lint api/openapi.yaml

golangci-lint:
ifeq (, $(shell which golangci-lint))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.31.0
endif
	golangci-lint run

lint: spectral golangci-lint

test:
	go test -v ./internal/...

verify: lint test

build: generate verify
	go build -o bin/app

mod:
	go mod tidy
	go mod verify

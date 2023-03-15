generate:
	go generate ./...

openapi-lint:
ifeq (, $(shell which vacuum))
	$(error vacuum binary is not found in path; install it from https://quobix.com/vacuum/installing/)
endif
	vacuum lint -d -n warn -r api/lint-rules.yaml api/openapi.yaml

go-lint:
ifeq (, $(shell which golangci-lint))
	$(error golangci-lint binary is not found in path; install it from https://golangci-lint.run/usage/install/)
endif
	golangci-lint run

lint: openapi-lint go-lint

test:
	go test -v -coverprofile=bin/coverage.out $$(go list ./pkg/business/... | grep -v /mock | grep -v /entity)

run: generate
	go run main.go

e2e-tests: generate
	go test -v ./tests/...

verify: lint test

assemble: generate
	go build -o bin/app main.go

build: assemble verify

mod:
	go mod tidy
	go mod verify

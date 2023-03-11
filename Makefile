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
	go test -v ./pkg/business/...

verify: lint test

assemble: generate
	go build -o bin/app

build: assemble verify
	go build -o bin/app

mod:
	go mod tidy
	go mod verify

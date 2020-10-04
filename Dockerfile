FROM golang:1.15-alpine3.12 AS build

WORKDIR /opt/app

RUN apk add make curl \
# remove spectral installation from Dockerfile
# when https://github.com/stoplightio/spectral/issues/1374 is fixed
 && apk add nodejs npm \
 && npm install -g @stoplight/spectral \
 && echo 'done'

COPY go.* ./

RUN go mod download \
 && echo 'done'

COPY api/openapi.yaml ./api/openapi.yaml
COPY configs/ ./configs/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY main.go ./
COPY Makefile ./

RUN CGO_ENABLED=0 make build \
 && echo 'done'

FROM alpine:3.12 AS run

WORKDIR /opt/app

COPY --from=build /opt/app/bin/app ./

ENTRYPOINT ["./app"]

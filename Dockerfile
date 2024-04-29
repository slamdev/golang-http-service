# syntax=docker/dockerfile:1
FROM docker.io/alpine:3.18.2

ENTRYPOINT ["/opt/app"]

COPY --link bin/app /opt/app

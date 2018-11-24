########################################################################################################################
# Common build image
########################################################################################################################
FROM golang:1.11 AS build
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/github.com/kfirz/gitzup
COPY ./Gopkg.lock ./Gopkg.toml ./
RUN dep ensure -v -vendor-only
COPY ./Makefile ./


########################################################################################################################
# Common alpine runtime image
# NOTES:
#   - installing "ca-certificates" & calling "update-ca-certificates" so Pub/Sub won't hang when connecting via 443/SSL
########################################################################################################################
FROM alpine:3.8 AS alpine
RUN apk update && apk add bash ca-certificates && update-ca-certificates && rm -rf /var/cache/apk/*


########################################################################################################################
# Agent builder & runtime
########################################################################################################################

# Build image
FROM build AS build-agent
COPY ./cmd ./cmd/
COPY ./internal ./internal/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 make agent

# Runtime image
FROM alpine AS agent
WORKDIR /app
COPY --from=build-agent /go/src/github.com/kfirz/gitzup/agent .
ENTRYPOINT ["/app/agent"]


########################################################################################################################
# API server builder
########################################################################################################################

# Build image
FROM build AS build-apiserver
COPY ./cmd ./cmd/
COPY ./internal ./internal/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 make apiserver

# Runtime image
FROM alpine AS apiserver
WORKDIR /app
COPY --from=build-apiserver /go/src/github.com/kfirz/gitzup/apiserver .
ENTRYPOINT ["/app/apiserver"]


########################################################################################################################
# Console builder
########################################################################################################################

# Build image
FROM build AS build-console
COPY ./cmd ./cmd/
COPY ./internal ./internal/
COPY ./web ./web/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 make console

# Runtime image
FROM alpine AS console
WORKDIR /app
COPY --from=build-console /go/src/github.com/kfirz/gitzup/console .
ENTRYPOINT ["/app/console"]


########################################################################################################################
# GCP builder
########################################################################################################################

# Build image
FROM build AS build-gcp
COPY ./cmd ./cmd/
COPY ./internal ./internal/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 make gcp

# Runtime image
FROM alpine AS gcp
WORKDIR /app
COPY --from=build-gcp /go/src/github.com/kfirz/gitzup/gcp .
ENTRYPOINT ["/app/gcp"]

# project image
FROM gcp AS gcp-project
ENTRYPOINT ["/app/gcp","project"]
CMD ["init"]

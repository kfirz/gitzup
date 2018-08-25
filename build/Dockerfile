FROM golang:1.9.7 AS build
ARG target
WORKDIR /go/src/github.com/kfirz/gitzup/
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY ./Gopkg.lock ./Gopkg.toml ./
RUN dep ensure -v -vendor-only
COPY ./api ./api/
COPY ./cmd ./cmd/
COPY ./internal ./internal/
COPY ./web ./web/
COPY ./Makefile ./
RUN make ${target}

########################################################################################################################
# NOTES:
#   - installing "ca-certificates" & calling "update-ca-certificates" so Pub/Sub won't hang when connecting via 443/SSL
########################################################################################################################

FROM ubuntu:18.04
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/* && \
    update-ca-certificates
WORKDIR /app
COPY --from=build /go/src/github.com/kfirz/gitzup/${target} .
ENTRYPOINT ["/app/${target}"]
CMD ["--help"]
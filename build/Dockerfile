FROM golang:1.9.7 AS build
WORKDIR /go/src/github.com/kfirz/gitzup/
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY ./Gopkg.lock ./Gopkg.toml ./
RUN dep ensure -v -vendor-only
ARG target
RUN mkdir -p ./cmd/api-server ./cmd/buildagent ./cmd/webhooks-server ./cmd/resources
COPY ./api ./api/
COPY ./cmd/${target} ./cmd/${target}
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
ARG target
COPY --from=build /go/src/github.com/kfirz/gitzup/${target} .
RUN ln -s /app/${target} /app/start && chmod +x /app/${target}
ENTRYPOINT ["/app/start"]
CMD ["--help"]

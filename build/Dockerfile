FROM golang:1.11.4 AS builder
WORKDIR /go/src/github.com/kfirz/gitzup
COPY ./cmd/ cmd/
COPY ./internal/ internal/
COPY ./pkg/ pkg/
COPY ./vendor/ vendor/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/gitzup ./cmd/manager

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/github.com/kfirz/gitzup/bin/gitzup /gitzup
ENTRYPOINT ["/gitzup"]

LABEL author=arik@kfirs.com
LABEL maintainer=arik@kfirs.com

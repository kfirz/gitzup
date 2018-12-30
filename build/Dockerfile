FROM golang:1.11 AS builder
MAINTAINER arik@kfirs.com
WORKDIR $GOPATH/src/github.com/kfirz/gitzup
COPY ./cmd/ cmd/
COPY ./internal/ internal/
COPY ./pkg/ pkg/
COPY ./vendor/ vendor/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix nocgo -o ./bin/manager ./cmd/manager

# final image
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/github.com/kfirz/gitzup/bin/manager /app
ENTRYPOINT ["/app"]

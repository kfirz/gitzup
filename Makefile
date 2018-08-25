CMD_AGENT = $(shell ls ./cmd/)

all: $(CMD_AGENT)

.PHONY: clean
clean:
	go clean
	rm -vf ./agent ./api-server ./webhooks-server

.PHONY: deps
deps:
	dep ensure

.PHONY: assets
assets:
	../../../../bin/go-bindata -o ./internal/assets/data.go -pkg assets -prefix api/ ./api

.PHONY: $(CMD_AGENT)
$(CMD_AGENT):
	go build -o $@ ./cmd/$@/

.PHONY: test
test:
	go test ./...

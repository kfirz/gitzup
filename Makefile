CMD_AGENT = $(shell ls ./cmd/)

all: $(CMD_AGENT)

.PHONY: clean
clean:
	go clean
	rm -vf ./agent ./api-server ./webhooks-server

.PHONY: $(CMD_AGENT)
$(CMD_AGENT):
	dep ensure
	go build -o $@ ./cmd/$@/

.PHONY: test
test:
	go test ./...

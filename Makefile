CMD_AGENT = $(shell ls ./cmd/)

.PHONY: clean
clean:
	go clean
	rm -f ./agent

.PHONY: $(CMD_AGENT)
$(CMD_AGENT):
	dep ensure
	go build -v -o $@ ./cmd/$@/

.PHONY: test
test:
	go test -v ./...

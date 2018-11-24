AGENT_ASSET_FILES = $(shell find ./internal/agent/assets -type f ! -name '*.go')
AGENT_ASSET_DIRS = $(shell find ./internal/agent/assets -type d)
COMMON_ASSET_FILES = $(shell find ./internal/common/assets -type f ! -name '*.go')
COMMON_ASSET_DIRS = $(shell find ./internal/common/assets -type d)
GCP_ASSET_FILES = $(shell find ./internal/gcp/assets -type f ! -name '*.go')
GCP_ASSET_DIRS = $(shell find ./internal/gcp/assets -type d)
TAG ?= dev

.PHONY: build
build: agent apiserver console gcp

.PHONY: clean
clean:
	rm -vf agent apiserver console gcp

internal/agent/assets/assets.go: $(AGENT_ASSET_FILES)
	go-bindata -o internal/agent/assets/assets.go -pkg assets $(AGENT_ASSET_DIRS)

internal/common/assets/assets.go: $(COMMON_ASSET_FILES)
	go-bindata -o internal/common/assets/assets.go -pkg assets $(COMMON_ASSET_DIRS)

internal/gcp/assets/assets.go: $(GCP_ASSET_FILES)
	go-bindata -o internal/gcp/assets/assets.go -pkg assets $(GCP_ASSET_DIRS)

# TODO: add golint

agent: ./cmd/agent/main.go $(SRC) internal/agent/assets/assets.go internal/common/assets/assets.go
	go build -o agent ./cmd/agent/main.go

apiserver: ./cmd/apiserver/main.go $(SRC) internal/common/assets/assets.go
	go build -o apiserver ./cmd/apiserver/main.go

console: ./cmd/console/main.go $(SRC) internal/common/assets/assets.go
	go build -o console ./cmd/console/main.go

gcp: ./cmd/gcp/main.go $(SRC) internal/gcp/assets/assets.go internal/common/assets/assets.go
	go build -o gcp ./cmd/gcp/main.go

.PHONY: test
test: build
	go test ./...

.PHONY: docker
docker:
	docker build -t gitzup/agent:$(TAG) -f ./build/Dockerfile --target agent .
	docker build -t gitzup/apiserver:$(TAG) -f ./build/Dockerfile --target apiserver .
	docker build -t gitzup/console:$(TAG) -f ./build/Dockerfile --target console .
	docker build -t gitzup/gcp:$(TAG) -f ./build/Dockerfile --target gcp .
	docker build -t gitzup/gcp-project:$(TAG) -f ./build/Dockerfile --target gcp-project .
	[[ "${PUSH}" == "true" ]] && \
		docker push gitzup/agent:$(TAG) && \
		docker push gitzup/apiserver:$(TAG) && \
		docker push gitzup/console:$(TAG) && \
		docker push gitzup/gcp:$(TAG) && \
		docker push gitzup/gcp-project:$(TAG) \
	|| true

.PHONY: latest
latest: docker
	docker tag gitzup/agent:$(TAG) gitzup/agent:latest
	docker tag gitzup/apiserver:$(TAG) gitzup/apiserver:latest
	docker tag gitzup/console:$(TAG) gitzup/console:latest
	docker tag gitzup/gcp:$(TAG) gitzup/gcp:latest
	docker tag gitzup/gcp-project:$(TAG) gitzup/gcp-project:latest
	[[ "${PUSH}" == "true" ]] && \
		docker push gitzup/agent:latest && \
		docker push gitzup/apiserver:latest && \
		docker push gitzup/console:latest && \
		docker push gitzup/gcp:latest && \
		docker push gitzup/gcp-project:latest \
	|| true

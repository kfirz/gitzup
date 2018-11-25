AGENT_ASSET_FILES = $(shell find ./internal/agent/assets -type f ! -name '*.go')
AGENT_ASSET_DIRS = $(shell find ./internal/agent/assets -type d)
COMMON_ASSET_FILES = $(shell find ./internal/common/assets -type f ! -name '*.go')
COMMON_ASSET_DIRS = $(shell find ./internal/common/assets -type d)
GCP_ASSET_FILES = $(shell find ./internal/gcp/assets -type f ! -name '*.go')
GCP_ASSET_DIRS = $(shell find ./internal/gcp/assets -type d)
REPO ?= gitzup
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
	docker build -t $(REPO)/agent:$(TAG) -f ./build/Dockerfile --target agent .
	docker build -t $(REPO)/apiserver:$(TAG) -f ./build/Dockerfile --target apiserver .
	docker build -t $(REPO)/console:$(TAG) -f ./build/Dockerfile --target console .
	docker build -t $(REPO)/gcp:$(TAG) -f ./build/Dockerfile --target gcp .
	docker build -t $(REPO)/gcp-project:$(TAG) -f ./build/Dockerfile --target gcp-project .

.PHONY: push-docker
push-docker: docker
	docker push $(REPO)/agent:$(TAG)
	docker push $(REPO)/apiserver:$(TAG)
	docker push $(REPO)/console:$(TAG)
	docker push $(REPO)/gcp:$(TAG)
	docker push $(REPO)/gcp-project:$(TAG)

.PHONY: latest
latest: docker
	docker tag $(REPO)/agent:$(TAG) $(REPO)/agent:latest
	docker tag $(REPO)/apiserver:$(TAG) $(REPO)/apiserver:latest
	docker tag $(REPO)/console:$(TAG) $(REPO)/console:latest
	docker tag $(REPO)/gcp:$(TAG) $(REPO)/gcp:latest
	docker tag $(REPO)/gcp-project:$(TAG) $(REPO)/gcp-project:latest

.PHONY: push-latest
push-latest: latest
	docker push $(REPO)/agent:latest
	docker push $(REPO)/apiserver:latest
	docker push $(REPO)/console:latest
	docker push $(REPO)/gcp:latest
	docker push $(REPO)/gcp-project:latest

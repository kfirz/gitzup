TAG ?= dev

# capture all asset files & asset directories in variables
# files list is used as a dependency list for the "assets" target; asset dirs are the arguments to "go-bindata"
ASSET_FILES = $(shell find ./api -type f)
ASSET_DIRS = $(shell find ./api -type d)

# capture source files for gitzup executables (used as dependency lists of Make targets)
CMD_API_SERVER_SRC = $(shell find ./cmd/api-server -name '*.go' -type f)
CMD_BUILDAGENT_SRC = $(shell find ./cmd/buildagent -name '*.go' -type f)
CMD_RESOURCES_SRC = $(shell find ./cmd/resources -name '*.go' -type f)
CMD_WEBHOOKS_SERVER_SRC = $(shell find ./cmd/webhooks-server -name '*.go' -type f)

# manipulate the gitzup-resource executables source code file-names into generic resource names;
# used to create a separate target per resource
CMD_RESOURCE_NAMES = $(basename $(shell find ./cmd/resources -name '*.go' -type f | sed 's|^./cmd/resources/||'))

# capture internal source code files (used as dependency lists of Make targets)
INTERNAL_SRC = $(shell find ./internal -type f)

# default target that builds all executables
all: internal/assets/data.go api-server buildagent resources webhooks-server test

# removes built executables
.PHONY: clean
clean:
	rm -vf api-server buildagent $(CMD_RESOURCE_NAMES) webhooks-server

# builds the embedded assets file
internal/assets/data.go: $(ASSET_FILES)
	$(GOPATH)/bin/go-bindata -o ./internal/assets/data.go -pkg assets -prefix api/ $(ASSET_DIRS)

# builds the API server executable
api-server: internal/assets/data.go $(INTERNAL_SRC) $(CMD_API_SERVER_SRC)
	go build -o api-server ./cmd/api-server/

# builds the build agent executable
buildagent: internal/assets/data.go $(INTERNAL_SRC) $(CMD_BUILDAGENT_SRC)
	go build -o buildagent ./cmd/buildagent/

# builds each of the gitzup resources
$(CMD_RESOURCE_NAMES): internal/assets/data.go $(INTERNAL_SRC) $(CMD_RESOURCES_SRC)
	go build -o $@ ./cmd/resources/$@.go

.PHONY: resources
resources: $(CMD_RESOURCE_NAMES)

# builds the webhooks server executable
webhooks-server: internal/assets/data.go $(INTERNAL_SRC) $(CMD_WEBHOOKS_SERVER_SRC)
	go build -o webhooks-server ./cmd/webhooks-server/

.PHONY: test
test: api-server buildagent webhooks-server $(CMD_RESOURCE_NAMES)
	go test ./...

# TODO: add golint

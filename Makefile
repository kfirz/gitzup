all: test manager

dep:
	dep ensure -v

generate:
	go generate ./pkg/... ./cmd/...

fmt:
	go fmt ./pkg/... ./cmd/...

vet:
	go vet ./pkg/... ./cmd/...

manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

manager: generate fmt vet
	go build -o bin/manager github.com/kfirz/gitzup/cmd/manager

deploy-crds: manifests
	./build/manifest_crds.sh | kubectl apply -f -

minikube: generate fmt vet manifests
	./build/minikube.sh

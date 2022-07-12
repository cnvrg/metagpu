#rsync -r /Users/dima/.go/src/github.com/AccessibleAI/metagpu-device-plugin/docs/* rancher@212.199.86.38:/tmp/docs

build:
	go build -ldflags="-X 'main.Build=$$(git rev-parse --short HEAD)' -X 'main.Version=1.0.0'" -v -o bin/mgdp cmd/mgdp/main.go

build-exporter:
	go build -ldflags="-X 'main.Build=$$(git rev-parse --short HEAD)' -X 'main.Version=1.0.0'" -v -o bin/mgex cmd/mgex/*.go

remote-sync:
	kubectl cp ./ $(shell kubectl get pods -lapp=dev-metagpu -A -ojson | jq -r '.items[] | .metadata.namespace + "/" + .metadata.name'):/opt/workdir/.go/github.com/metagpu

remote-debug:
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient  ./cmd/mgdp/main.go -- start

docker-dev-build:
	docker buildx build --platform linux/amd64 --push -t cnvrg/golang-dvl:latest -f Dockerfile.dev .

docker-build: build-proto
	docker build \
	 --platform linux/x86_64 \
     --build-arg BUILD_SHA=$(shell git rev-parse --short HEAD) \
     --build-arg BUILD_VERSION=1.0.0 \
     -t docker.io/cnvrg/metagpu-device-plugin:$(shell git rev-parse --abbrev-ref HEAD) .

build-mgctl:
	go build -ldflags="-X 'main.Build=$$(git rev-parse --short HEAD)' -X 'main.Version=1.0.0'" -v -o bin/mgctl cmd/mgctl/*.go

docker-push:
	docker push docker.io/cnvrg/metagpu-device-plugin:$(shell git rev-parse --abbrev-ref HEAD)

build-proto:
	buf mod update pkg/mgsrv/deviceapi
	buf lint
	buf build
	buf generate

generate-manifests:
	helm template chart/ -n cnvrg --set tag=$(shell git rev-parse --abbrev-ref HEAD) > deploy/static.yaml

.PHONY: deploy
deploy:
	helm template chart/ --set tag=$(shell git rev-parse --abbrev-ref HEAD) | kubectl apply -f -

test-all:
	go test ./pkg/... -v

test-allocator:
	go test ./pkg/allocator/... -v

test-gpumgr:
	go test ./pkg/gpumgr/... -v
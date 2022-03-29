#rsync -r /Users/dima/.go/src/github.com/AccessibleAI/metagpu-device-plugin/docs/* rancher@212.199.86.38:/tmp/docs

build:
	go build -ldflags="-X 'main.Build=$$(git rev-parse --short HEAD)' -X 'main.Version=0.1.1'" -v -o bin/mgdp cmd/mgdp/main.go

build-exporter:
	go build -ldflags="-X 'main.Build=$$(git rev-parse --short HEAD)' -X 'main.Version=0.1.1'" -v -o bin/mgex cmd/mgex/*.go

debug-remote:
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient  ./cmd/mgdp/main.go -- start

docker-build: build-proto
	docker build \
	 --platform linux/x86_64 \
     --build-arg BUILD_SHA=$(shell git rev-parse --short HEAD) \
     --build-arg BUILD_VERSION=0.0.1 \
     -t docker.io/cnvrg/metagpu-device-plugin:$(shell git rev-parse --abbrev-ref HEAD) .

build-mgctl:
	go build -ldflags="-X 'main.Build=$$(git rev-parse --short HEAD)' -X 'main.Version=0.1.1'" -v -o bin/mgctl cmd/mgctl/*.go

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

dev-sync-azure:
	rsync -av  --exclude 'bin' --exclude '.git'  /Users/dima/.go/src/github.com/AccessibleAI/metagpu-device-plugin/* root@20.120.94.51:/root/.go/src/github.com/AccessibleAI/metagpu-device-plugin

dev-sync-trex:
	rsync -av  --exclude 'bin' --exclude '.git'  /Users/dima/.go/src/github.com/AccessibleAI/metagpu-device-plugin/* root@212.199.86.38:/root/.go/src/github.com/AccessibleAI/metagpu-device-plugin

test-all:
	go test ./pkg/... -v

test-allocator:
	go test ./pkg/allocator/... -v

test-gpumgr:
	go test ./pkg/gpumgr/... -v
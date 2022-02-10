build-mac:
	go build -ldflags="-X 'main.Build=$$(git rev-parse --short HEAD)' -X 'main.Version=0.1.1'" -v -o bin/metagpu-dp-darwin-x86_64 main.go

debug-remote:
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient  ./cmd/metagpu-device-plugin/main.go -- start

docker-build: build-proto
	docker build \
	 --platform linux/x86_64 \
     --build-arg BUILD_SHA=$(shell git rev-parse --short HEAD) \
     --build-arg BUILD_VERSION=0.0.1 \
     -t docker.io/cnvrg/metagpu-device-plugin:latest .

docker-push:
	docker push docker.io/cnvrg/metagpu-device-plugin:latest

controller-generate:
	 controller-gen-v0.8.0 object paths=./cmd/metagpu-controller/api/...

controller-manifests:
	controller-gen-v0.8.0 crd paths=./cmd/metagpu-controller/api/... output:artifacts:config=./config/crd/bases

build-proto:
	buf mod update pkg/metagpusrv/deviceapi
	buf lint
	buf build
	buf generate
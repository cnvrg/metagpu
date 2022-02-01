build-mac:
	go build -ldflags="-X 'main.Build=$$(git rev-parse --short HEAD)' -X 'main.Version=0.1.1'" -v -o bin/fractor-darwin-x86_64 main.go

debug-remote:
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient  ./main.go -- start

docker-build:
	docker build \
     --build-arg BUILD_SHA=$(shell git rev-parse --short HEAD) \
     --build-arg BUILD_VERSION=0.0.1 \
     -t docker.io/cnvrg/fractor:latest .

docker-push:
	docker push docker.io/cnvrg/fractor:latest

FROM golang:1.17.3 as builder
ARG BUILD_SHA
ARG BUILD_VERSION
WORKDIR /root/.go/src/fractor
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY cmd cmd
COPY pkg pkg
RUN CGO_LDFLAGS_ALLOW='-Wl,--unresolved-symbols=ignore-in-object-files' \
    go build \
    -ldflags="-s -w -X 'main.Build=${BUILD_SHA}' -X 'main.Version=${BUILD_VERSION}'" \
    -o fractor cmd/fractor/main.go

FROM ubuntu:20.04
WORKDIR /opt/app-root
COPY --from=builder /root/.go/src/fractor/fractor /opt/app-root/fractor

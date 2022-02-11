FROM golang:1.17.3 as builder
ARG BUILD_SHA
ARG BUILD_VERSION
WORKDIR /root/.go/src/metagpu
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY cmd cmd
COPY pkg pkg
COPY gen gen
RUN CGO_LDFLAGS_ALLOW='-Wl,--unresolved-symbols=ignore-in-object-files' \
    go build \
    -ldflags="-s -w -X 'main.Build=${BUILD_SHA}' -X 'main.Version=${BUILD_VERSION}'" \
    -o metagpu-device-plugin cmd/metagpu-device-plugin/main.go

FROM nvidia/cuda:11.6.0-base-ubuntu20.04


ENV NVIDIA_DISABLE_REQUIRE="true"
ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=utility

LABEL io.k8s.display-name="cnvrg.io Meta GPU Device Plugin"
LABEL name="cnvrg.io Device Plugin"
LABEL vendor="cnvrg.io"
ARG PLUGIN_VERSION="N/A"
LABEL version=${PLUGIN_VERSION}
LABEL release="N/A"
LABEL summary="cnvrg.io device plugin for Kubernetes"
LABEL description="See summary"
RUN apt update -y \
    && apt install -y vim
COPY --from=builder /root/.go/src/metagpu/metagpu-device-plugin /usr/bin/metagpu-device-plugin
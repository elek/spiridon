VERSION 0.6

build:
   FROM golang
   ENV CGO_ENABLED=0
   WORKDIR /go/work
   COPY . .
   RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod go install
   SAVE ARTIFACT /go/bin binary
build-image:
   FROM alpine
   COPY +build/binary /usr/local/bin/
   SAVE IMAGE --push ghcr.io/elek/spiridon

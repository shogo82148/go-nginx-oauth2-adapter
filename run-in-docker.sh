#!/bin/sh

CURRENT=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$CURRENT/.mod"
docker run --rm -it \
    -e GO111MODULE=on \
    -e CGO_ENABLED=0 \
    -v "$CURRENT/.mod":/go/pkg/mod \
    -v "$CURRENT":/go/src/github.com/shogo82148/go-nginx-oauth2-adapter \
    -w /go/src/github.com/shogo82148/go-nginx-oauth2-adapter golang:1.16.0 "$@"

#!/bin/sh

CURRENT=$(cd "$(dirname "$0")" && pwd)
docker run --rm -it \
    -v "$CURRENT":/go/src/github.com/shogo82148/go-nginx-oauth2-adapter \
    -w /go/src/github.com/shogo82148/go-nginx-oauth2-adapter golang:1.11.1 "$@"

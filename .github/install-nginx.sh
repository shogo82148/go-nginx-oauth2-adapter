#!/bin/bash

set -eux

NGINX_VERSION=1.19.1
CURRENT=$(cd "$(dirname "$0")" && pwd)

if [[ ! -d "$CURRENT/tmp/nginx-$NGINX_VERSION" ]]; then
    mkdir -p "$CURRENT/tmp"
    cd "$CURRENT/tmp"
    curl -OL "https://nginx.org/download/nginx-$NGINX_VERSION.tar.gz"
    tar xzf "nginx-$NGINX_VERSION.tar.gz"
fi

rm -rf "$CURRENT/nginx"
cd "$CURRENT/tmp/nginx-$NGINX_VERSION"
./configure --prefix="$CURRENT/nginx"
make
make install

echo "::add-path::$CURRENT/nginx/sbin"

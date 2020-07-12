#!/bin/bash

set -eux

CURRENT=$(cd "$(dirname "$0")" && pwd)
H2O_VERSION=2.2.6

echo "::add-path::$CURRENT/h2o/bin"

if "$CURRENT/h2o/bin/h2o" --version | grep -F "h2o version $H2O_VERSION"; then
    : "h2o version $H2O_VERSION is already installed. nothing to do."
    exit 0
fi

if [[ ! -d "$CURRENT/tmp/h2o-$H2O_VERSION" ]]; then
    mkdir -p "$CURRENT/tmp"
    cd "$CURRENT/tmp"
    curl -L "https://github.com/h2o/h2o/archive/v$H2O_VERSION.tar.gz" -o "h2o-$H2O_VERSION.tar.gz"
    tar xzf "h2o-$H2O_VERSION.tar.gz"
fi

rm -rf "$CURRENT/h2o"
cd "$CURRENT/tmp/h2o-$H2O_VERSION"
cmake -DWITH_BUNDLED_SSL=on -DWITH_MRUBY=on -DCMAKE_INSTALL_PREFIX="$CURRENT/h2o" .
make
make install

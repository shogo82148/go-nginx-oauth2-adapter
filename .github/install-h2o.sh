#!/bin/bash

set -eux

CURRENT=$(cd "$(dirname "$0")" && pwd)
H2O_VERSION=2.2.6

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

echo "::add-path::$CURRENT/h2o/bin"

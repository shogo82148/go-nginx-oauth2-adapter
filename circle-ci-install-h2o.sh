#!/bin/bash

set -x
set -e

H2O_VERSION=2.2.5

if [[ ! -d "$HOME/h2o-$H2O_VERSION" ]]; then
    cd ~/
    curl -OL "https://github.com/h2o/h2o/archive/v$H2O_VERSION.tar.gz"
    tar xzf "v$H2O_VERSION.tar.gz"
fi

cd "$HOME/h2o-$H2O_VERSION"
cmake -DWITH_BUNDLED_SSL=on -DWITH_MRUBY=on .
make
sudo make install

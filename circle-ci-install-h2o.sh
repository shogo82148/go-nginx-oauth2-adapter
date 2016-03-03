#!/bin/bash

set -x
set -e

if [[ ! -d "$HOME/h2o-1.7.0" ]]; then
    cd ~/
    curl -OL https://github.com/h2o/h2o/archive/v1.7.0.tar.gz
    tar xzf v1.7.0.tar.gz
fi

cd ~/h2o-1.7.0
cmake -DWITH_BUNDLED_SSL=on .
make
sudo make install

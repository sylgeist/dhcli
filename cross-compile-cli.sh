#!/bin/bash

set -euo pipefail

BINARY="bin/dhcli"

if [ ! -x "$(which packingslip)" ]; then
    echo "packingslip binary not found" 1>&2
    exit 1
fi

do_build() {
    export GOOS="$1"
    export GOARCH="$2"
    ARTIFACTS="artifacts/$GOOS/$GOARCH"
    mkdir -p "$ARTIFACTS"
    echo "Building for $GOOS on $GOARCH..."
    make clean && make "$BINARY" && mv bin/* "$ARTIFACTS"
}

rm -rf artifacts

do_build linux amd64
do_build linux arm64
do_build windows amd64
do_build darwin amd64
do_build darwin arm64

unset GOOS
unset GOARCH
make clean

cd artifacts
packingslip version

echo "Writing artifacts/packing_slip.json..."
find . -type f | sed 's/^\.\///' | grep -v 'packing_slip\.json' | \
    xargs packingslip create \
    --version "$(git rev-parse HEAD)" \
    > packing_slip.json


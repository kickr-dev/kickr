#!/bin/sh

install-tool golang "$(grep '^toolchain ' go.mod | awk '{print $2}' | sed 's/^go//')"
make testdata

make build
./kickr

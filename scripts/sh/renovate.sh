#!/bin/sh

install-tool golang "$(grep '^toolchain ' go.mod | awk '{print $2}' | sed 's/^go//')"
TESTDATA=1 go test ./... -run ^TestGenerate_ -count 1 -timeout=15s
go run ./cmd/kickr

#!/bin/sh

set -e

BUILD="-X=main.build=$(git rev-parse HEAD) -X=main.version=$(git describe --tags --abbrev=0)"

set -x
go test ./... -cover
go build -tags "netgo osusergo" -ldflags "$BUILD" -o jrnl.out
set +x

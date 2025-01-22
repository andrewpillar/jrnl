

GOPATH := $(shell go env GOPATH)

VERSION := $(shell git describe --tags)
BUILD   := $(shell git log -n 1 --format='format:%h %cd' HEAD)

TAGS    := netgo,osusergo
LDFLAGS := -s -w -X 'main.version=$(VERSION)' -X 'main.build=$(BUILD)'

CMD := jrnl
BIN := bin

BUILD_DEPS := clean mod generate fmt test

all: build

mod:
	go mod tidy

generate:
	go generate ./...

fmt:
	gofmt -s -w .

test:
	go test -v -cover ./...

build: $(BUILD_DEPS)
	mkdir -p $(BIN)
	go build -trimpath -ldflags "$(LDFLAGS)" -tags "$(TAGS)" -o $(BIN)/$(CMD) .

install: build
	cp $(BIN)/jrnl $(GOPATH)/bin/jrnl

clean:
	go clean -testcache
	rm -rf $(BIN)/
	rm -rf testdata/_* testdata/.cache/ testdata/remote/ testdata/jrnl.yml

.PHONY = all clean fmt generate test

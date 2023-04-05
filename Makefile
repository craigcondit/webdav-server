BASE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: all
all:
	$(MAKE) -C $(dir (BASE_DIR)) build

.PHOHNY: build
build: bin/webdav-server

.PHONY: clean
clean:
	@rm -rf bin sandbox
	go clean -cache -testcache -r

sandbox:
	@mkdir -p sandbox

bin/webdav-server: sandbox cmd/**/*.go pkg/**/*.go
	CGO_ENABLED=0 \
	go build -a -ldflags '-extldflags "-static"' -tags netgo -installsuffix netgo \
	-o bin/webdav-server ./cmd/webdav

.PHONY:
run: build
	bin/webdav-server

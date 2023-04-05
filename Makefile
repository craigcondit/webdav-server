BASE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

ifeq ($(VERSION),)
VERSION := latest
endif

ifeq ($(REGISTRY),)
REGISTRY=craigcondit
endif

ifeq ($(HOST_ARCH),)
HOST_ARCH := $(shell uname -m)
endif
ifeq (x86_64, $(HOST_ARCH))
EXEC_ARCH := amd64
DOCKER_ARCH := amd64
else ifeq (i386, $(HOST_ARCH))
EXEC_ARCH := 386
DOCKER_ARCH := i386
else ifneq (,$(filter $(HOST_ARCH), arm64 aarch64))
EXEC_ARCH := arm64
DOCKER_ARCH := arm64v8
else ifeq (armv7l, $(HOST_ARCH))
EXEC_ARCH := arm
DOCKER_ARCH := arm32v7
else
$(info Unknown architecture "${HOST_ARCH}" defaulting to: amd64)
EXEC_ARCH := amd64
DOCKER_ARCH := amd64
endif

.PHONY: all
all:
	$(MAKE) -C $(dir (BASE_DIR)) build image

.PHONY: build
build: bin/webdav-server.dev

.PHONY: build-release
build-release: bin/webdav-server

.PHONY: image
image: build-release Dockerfile
	rm -rf .tmp/docker-dirs
	mkdir -p .tmp/docker-dirs/root/conf .tmp/docker-dirs/user/content
	DOCKER_BUILDKIT=1 \
	docker build . -t "${REGISTRY}/webdav-server-${DOCKER_ARCH}:${VERSION}" \
	--platform "linux/${DOCKER_ARCH}"
	
.PHONY: clean
clean:
	@rm -rf bin sandbox .tmp
	go clean -cache -testcache -r

sandbox:
	@mkdir -p sandbox

bin/webdav-server.dev: sandbox cmd/**/*.go pkg/**/*.go
	CGO_ENABLED=0 \
	go build -a -ldflags '-extldflags "-static"' -tags netgo -installsuffix netgo \
	-o bin/webdav-server.dev ./cmd/webdav

bin/webdav-server: sandbox cmd/**/*.go pkg/**/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH="${EXEC_ARCH}" \
	go build -a -ldflags '-extldflags "-static"' -tags netgo -installsuffix netgo \
	-o bin/webdav-server ./cmd/webdav
	
.PHONY:
run: build
	bin/webdav-server.dev

run-docker: image
	docker run -it --rm=true \
	-p 8080:8080 \
	-v "$(BASE_DIR)/sandbox:/content" \
	-v "$(BASE_DIR)/conf:/conf" \
	"${REGISTRY}/webdav-server-${DOCKER_ARCH}:${VERSION}"

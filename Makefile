.PHONY: clean run deploy build.local build.linux build.docker deploy

BINARY        ?= geddit
SOURCES       = $(shell find . -name '*.go')
STATICS       = $(shell find tmpl -name '*.*')
VERSION       := $(shell date '+%Y%m%d%H%M%S')
IMAGE         ?= deploy.glv.one/pitr/$(BINARY)
DOCKERFILE    ?= Dockerfile
BUILD_FLAGS   ?= -v
LDFLAGS       ?= -w -s

default: run

clean:
	rm -rf build

run: build.local
	./build/$(BINARY)

build.local: build/$(BINARY)
build.linux: build/linux/$(BINARY)

build/$(BINARY): $(SOURCES)
	CGO_ENABLED=1 go build -o build/$(BINARY) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" .

build/linux/$(BINARY): $(SOURCES)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build $(BUILD_FLAGS) -o build/linux/$(BINARY) -ldflags "$(LDFLAGS)" .

build.docker:
	docker build --rm -t "$(IMAGE):$(VERSION)" -f $(DOCKERFILE) .

deploy: build.docker
	docker push "$(IMAGE):$(VERSION)"

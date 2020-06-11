.PHONY: clean run deploy build.local build.linux

BINARY        ?= gemininews
SOURCES       = $(shell find . -name '*.go') assets.go
STATICS       = $(shell find static -name '*.*')
BUILD_FLAGS   ?= -v
PORT          ?= 1965
LDFLAGS       ?= -w -s -X main.port=$(PORT)

default: run

clean:
	rm -rf build

run: build.local
	./build/$(BINARY)

deploy: build.linux
	scp build/linux/$(BINARY) ec2-user@$(PRODUCTION):$(BINARY)-next
	ssh ec2-user@$(PRODUCTION) 'cp $(BINARY) $(BINARY)-old'
	ssh ec2-user@$(PRODUCTION) 'mv $(BINARY)-next $(BINARY)'
	ssh ec2-user@$(PRODUCTION) 'sudo systemctl restart $(BINARY)'

rollback:
	ssh ec2-user@$(PRODUCTION) 'mv $(BINARY)-old $(BINARY)'
	ssh ec2-user@$(PRODUCTION) 'sudo systemctl restart $(BINARY)'

assets.go: $(STATICS)
	go run cmd/build_assets.go $(STATICS)

build.local: build/$(BINARY)
build.linux: build/linux/$(BINARY)

build/$(BINARY): $(SOURCES)
	CGO_ENABLED=0 go build -o build/$(BINARY) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" .

build/linux/$(BINARY): $(SOURCES)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/linux/$(BINARY) -ldflags "$(LDFLAGS)" .

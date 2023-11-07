prefix ?= /usr

VERSION = $(shell git describe --always --long --dirty --tags)
LDFLAGS = "-X pulwar.isi.edu/sabres/orchestrator/pkg/common.Version=$(VERSION)"

#all: docker
all: clean code

protobuf: protobuf-inventory

code: build/iservice build/ictl 

build/ictl: inventory/cli/main.go | build protobuf-inventory
	go build -ldflags=$(LDFLAGS) -o $@ $<

build/iservice: inventory/service/main.go | build protobuf-inventory
	go build -ldflags=$(LDFLAGS) -o $@ $<

build:
	mkdir -p build

clean:
	rm -rf build

protobuf-inventory:
	protoc -I=inventory/protocol --go_out=inventory/protocol --go_opt=paths=source_relative \
		--go-grpc_out=inventory/protocol --go-grpc_opt=paths=source_relative  \
		 inventory/protocol/*.proto

test:
	go test -v ./...

REGISTRY ?= docker.io
REPO ?= isilincoln
TAG ?= latest
BUILD_ARGS ?= --no-cache

docker: $(REGISTRY)/$(REPO)/orchestrator-inventory

$(REGISTRY)/$(REPO)/orchestrator-inventory:
	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f inventory/Dockerfile -t $(@):$(TAG) .
	$(if ${PUSH},$(call docker-push))

define docker-push
	@docker push $(@):$(TAG)
endef

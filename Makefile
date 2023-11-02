prefix ?= /usr

VERSION = $(shell git describe --always --long --dirty --tags)
LDFLAGS = "-X pulwar.isi.edu/sabres/orchestrator/pkg/common.Version=$(VERSION)"

#all: docker
all: clean code

protobuf: protobuf-inventory

code: build/iservice #build/ictl 

build/ictl: cli/inventory/main.go | build protobuf-inventory
	go build -ldflags=$(LDFLAGS) -o $@ $<

build/iservice: inventory/main.go | build protobuf-inventory
	go build -ldflags=$(LDFLAGS) -o $@ $<

build:
	mkdir -p build

clean:
	rm -rf build

protobuf-inventory:
	protoc -I=inventory/service --go_out=inventory/service --go_opt=paths=source_relative \
		--go-grpc_out=inventory/service --go-grpc_opt=paths=source_relative  \
		 inventory/service/*.proto

test:
	go test -v ./...

#REGISTRY ?= docker.io
#REPO ?= isilincoln
#TAG ?= latest
#BUILD_ARGS ?= --no-cache
#
#docker: $(REGISTRY)/$(REPO)/orchestrator-api $(REGISTRY)/$(REPO)/orchestrator-watcher $(REGISTRY)/$(REPO)/sabres-client
#
#$(REGISTRY)/$(REPO)/orchestrator-api:
#	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f Dockerfile.api -t $(@):$(TAG) .
#	$(if ${PUSH},$(call docker-push))
#
#$(REGISTRY)/$(REPO)/orchestrator-watcher:
#	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f Dockerfile.watcher -t $(@):$(TAG) .
#	$(if ${PUSH},$(call docker-push))
#
#$(REGISTRY)/$(REPO)/sabres-client:
#	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f Dockerfile.sabres-client -t $(@):$(TAG) .
#	$(if ${PUSH},$(call docker-push))
#
#define docker-push
#	@docker push $(@):$(TAG)
#endef

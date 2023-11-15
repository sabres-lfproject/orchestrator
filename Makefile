prefix ?= /usr

VERSION = $(shell git describe --always --long --dirty --tags)
LDFLAGS = "-X pulwar.isi.edu/sabres/orchestrator/pkg/common.Version=$(VERSION)"

#all: docker
all: clean code mock

protobuf: protobuf-inventory

code: build/iservice build/ictl build/dservice build/dscanner build/dctl

mock: build/dmock

test:
	go test -v ./...

build/ictl: inventory/cli/main.go | build protobuf-inventory
	go build -ldflags=$(LDFLAGS) -o $@ $<

build/iservice: inventory/service/main.go | build protobuf-inventory
	go build -ldflags=$(LDFLAGS) -o $@ $<

build/dmock: discovery/mock/main.go
	go build -ldflags=$(LDFLAGS) -o $@ $<

build/dservice: discovery/service/main.go | build protobuf-discovery
	go build -ldflags=$(LDFLAGS) -o $@ $<

build/dscanner: discovery/scanner/main.go | build protobuf-inventory protobuf-discovery
	go build -ldflags=$(LDFLAGS) -o $@ $<

build/dctl: discovery/cli/main.go | build protobuf-discovery
	go build -ldflags=$(LDFLAGS) -o $@ $<

build:
	mkdir -p build

clean:
	rm -rf build

protobuf-inventory:
	protoc -I=inventory/protocol --go_out=inventory/protocol --go_opt=paths=source_relative \
		--go-grpc_out=inventory/protocol --go-grpc_opt=paths=source_relative  \
		 inventory/protocol/*.proto

protobuf-discovery:
	protoc -I=discovery/protocol --go_out=discovery/protocol --go_opt=paths=source_relative \
		--go-grpc_out=discovery/protocol --go-grpc_opt=paths=source_relative  \
		 discovery/protocol/*.proto

test:
	go test -v ./...

REGISTRY ?= docker.io
REPO ?= isilincoln
TAG ?= latest
#BUILD_ARGS ?= --no-cache

docker: $(REGISTRY)/$(REPO)/orchestrator-inventory $(REGISTRY)/$(REPO)/orchestrator-discovery $(REGISTRY)/$(REPO)/orchestrator-mock-discovery $(REGISTRY)/$(REPO)/orchestrator-discovery-scanner

$(REGISTRY)/$(REPO)/orchestrator-inventory:
	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f inventory/Dockerfile -t $(@):$(TAG) .
	$(if ${PUSH},$(call docker-push))

$(REGISTRY)/$(REPO)/orchestrator-discovery:
	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f discovery/Dockerfile -t $(@):$(TAG) .
	$(if ${PUSH},$(call docker-push))

$(REGISTRY)/$(REPO)/orchestrator-discovery-scanner:
	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f discovery/scanner/Dockerfile -t $(@):$(TAG) .
	$(if ${PUSH},$(call docker-push))

$(REGISTRY)/$(REPO)/orchestrator-mock-discovery:
	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f discovery/mock/Dockerfile -t $(@):$(TAG) .
	$(if ${PUSH},$(call docker-push))

define docker-push
	@docker push $(@):$(TAG)
endef

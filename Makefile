prefix ?= /usr

VERSION = $(shell git describe --always --long --dirty --tags)
LDFLAGS = "-X pulwar.isi.edu/sabres/orchestrator/pkg/common.Version=$(VERSION)"

#all: docker
all: clean code mock

protobuf: protobuf-inventory protobuf-discovery protobuf-networking protobuf-management

code: inventory discovery networking sdcore manager

mock: build/dmock

inventory: build/iservice build/ictl

discovery: build/dservice build/dscanner build/dctl

networking: build/snet build/snctl

sdcore: build/sdcli

manager: build/smgmt build/scli


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

build/snet: sabres/network/service/main.go | build protobuf-inventory protobuf-networking
	go build -ldflags=$(LDFLAGS) -o $@ $<

build/snctl: sabres/network/cli/main.go | build protobuf-networking
	go build -ldflags=$(LDFLAGS) -o $@ $<

build/sdcli: sabres/sd-core-interface/main.go
	go build -ldflags=$(LDFLAGS) -o $@ $<

build/smgmt: sabres/manager/service/main.go | build protobuf-inventory protobuf-networking protobuf-management
	go build -ldflags=$(LDFLAGS) -o $@ $<

build/sctl: sabres/manager/cli/main.go | build protobuf-management
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

protobuf-networking:
	protoc -I=sabres/network/protocol --go_out=sabres/network/protocol --go_opt=paths=source_relative \
		--go-grpc_out=sabres/network/protocol --go-grpc_opt=paths=source_relative  \
		 sabres/network/protocol/*.proto

protobuf-management:
	protoc -I=sabres/manager/protocol --go_out=sabres/manager/protocol --go_opt=paths=source_relative \
		--proto_path=sabres/ \
		--go-grpc_out=sabres/manager/protocol --go-grpc_opt=paths=source_relative  \
		 sabres/manager/protocol/*.proto

test:
	go test -v ./...

REGISTRY ?= docker.io
REPO ?= isilincoln
TAG ?= latest
#BUILD_ARGS ?= --no-cache

docker: $(REGISTRY)/$(REPO)/orchestrator-inventory-api $(REGISTRY)/$(REPO)/orchestrator-discovery-api $(REGISTRY)/$(REPO)/orchestrator-mock-discovery $(REGISTRY)/$(REPO)/orchestrator-discovery-scanner $(REGISTRY)/$(REPO)/orchestrator-sabres-network

$(REGISTRY)/$(REPO)/orchestrator-inventory-api:
	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f inventory/service/Dockerfile -t $(@):$(TAG) .
	$(if ${PUSH},$(call docker-push))

$(REGISTRY)/$(REPO)/orchestrator-discovery-api:
	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f discovery/service/Dockerfile -t $(@):$(TAG) .
	$(if ${PUSH},$(call docker-push))

$(REGISTRY)/$(REPO)/orchestrator-discovery-scanner:
	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f discovery/scanner/Dockerfile -t $(@):$(TAG) .
	$(if ${PUSH},$(call docker-push))

$(REGISTRY)/$(REPO)/orchestrator-mock-discovery:
	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f discovery/mock/Dockerfile -t $(@):$(TAG) .
	$(if ${PUSH},$(call docker-push))

$(REGISTRY)/$(REPO)/orchestrator-sabres-network:
	@docker build ${BUILD_ARGS} $(DOCKER_QUIET) -f sabres/network/service/Dockerfile -t $(@):$(TAG) .
	$(if ${PUSH},$(call docker-push))

define docker-push
	@docker push $(@):$(TAG)
endef

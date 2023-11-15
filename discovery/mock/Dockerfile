FROM golang:bullseye as build

RUN apt update && apt upgrade -qy
RUN apt install -y \
    build-essential \
    golang \
    make \
    ca-certificates \
    protobuf-compiler \
    vim

COPY . /orchestrator
WORKDIR /orchestrator

RUN rm -rf build

RUN mkdir -p /go
ENV GOPATH /go
ENV GOBIN /go/bin
ENV PATH "$PATH:$GOBIN"
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

RUN make mock

FROM ubuntu:22.04 as final

RUN apt update && apt install -qy \
    iproute2

COPY --from=build /orchestrator/build/* /usr/bin/

RUN mkdir /data

COPY --from=build /orchestrator/ /data
    
RUN mkdir /etc/orchestrator

ENV ETCDPORT 2379
ENV ETCDHOST localhost
ENV MOCKDISCOVERYPORT 15015
ENV DATADIR /data/discovery/mock/

# orchestratorctld grpc
EXPOSE ${MOCKDISCOVERYPORT}

CMD /usr/bin/dmock

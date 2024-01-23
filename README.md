# Sabres Orchestrator

The Sabres Orchestrator is a high level orchestrator that was developed to sit
on top of pre-existing orchestrators such as neph.io, onap, sd-core to manage
network resources in a security-concious manner.  Sabres is responsible for
inventorying network resources, generating network graphs, network slices,
proxy re-encryption, path validation, and network obfuscation.

The Sabres Orchestrator (SO) was developed for networks with non-uniform
ownership, where multiple entities control different parts of the network.
Some parts of the network may be more or less trusted than others, the SO
is responsible for tracking these constraints and building network slices
across untrusted resources and providing privacy and security guarentees to the
slice operators.

## Building

To build the code, the following dependencies are required:

```
golang-1.19
make
protobuf-compiler
```

The code can be built using `make` in the top-level directory.

### Containers

To build all containers use:

```
make docker
```

Note that the makefile calls docker, so if another subsystem is being used to
build containers, then that system will need to be used in place of docker.

The dockerfiles can be found in the service directory of each service. E.g.,

```
inventory/service/Dockerfile
```

## Deploying

To deploy the containers, there exists a helm directory which can be used to
deploy the subservices together into a single kubernetes deployment.

## Contributing

Any contributions via issues, requests, bug tickets, and/or pull requests are
welcome.

## Acknowledgements

This research is supported by DARPA under grant number HR001120C0157. The views, opinions, and/or findings expressed are those of the author(s) and should not be interpreted as representing the official views or policies of the sponsoring organizations, agencies, or the U.S. Government.

## License

The 3-Clause BSD License

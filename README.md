# EPON OLT adapter

The EPON OLT adapter connects the VOLTHA
core to an OLT device
running the OLT hardware to support IEEE standard PON (SIEPON* package B). 

*IEEE P1904.1 Service Interoperability in Ethernet Passive Optical Network
> NOTE: This adapter has been verified with the following combinations:
> * [VOLTHA 2.5](https://docs.voltha.org/voltha-2.5/release_notes/voltha_2.5.html) 
> * Tibit MicroPlug OLT (Firmware R1.3.0)
> * [Technology profile for EPON](https://github.com/opencord/voltha-lib-go/blob/master/pkg/techprofile/SingleQueueEponProfile.json)
> * MD5 is implemented as IEEE802.1X authentication. 

## Development `make` targets

The `Makefile` contains many commands that are useful in development:

```
build                     : Alias for 'docker build'
clean                     : Removes any local filesystem artifacts generated by a build
distclean                 : Removes any local filesystem artifacts generated by a build or test run
docker-build-profile      : Build eponolt adapter docker image with profiling enabled
docker-build              : Build eponolt adapter docker image
docker-kind-load          : Load docker images into a KinD cluster
docker-push               : Push the docker images to an external repository
help                      : Print help for each Makefile target
lint-dockerfile           : Perform static analysis on Dockerfile
lint-mod                  : Verify the Go dependencies
lint                      : Run all lint targets
local-lib-go              : Copies a local version of the voltha-lib-go dependency into the vendor directory
local-protos              : Copies a local version of the voltha-protos dependency into the vendor directory
mod-update                : Update go mod files
sca                       : Runs static code analysis with the golangci-lint tool
test                      : Run unit tests
```

Some highlights:

- It's recommended that you run the `lint`, `sca`, and `test` targets before
  submitting code changes.

- The `docker-*` targets for building and pushing Docker images depend on the
  variables `DOCKER_REGISTRY`, `DOCKER_REPOSITORY`, and `DOCKER_TAG` as
  [described in the CORD
  documentation](https://guide.opencord.org/developer/test_release_software.html#publish-docker-container-images-to-public-dockerhub-job-docker-publish)

- If you make changes the dependencies in the `go.mod` file, you will need to
  run `make mod-update` to update the `go.sum` and `vendor` directory.

### Building with a Local Copy of `voltha-protos` or `voltha-lib-go`

If you want to build/test using a local copy of the `voltha-protos` or
`voltha-lib-go` libraries this can be accomplished by using the environment
variables `LOCAL_PROTOS` and `LOCAL_LIB_GO`. These environment variables should
be set to the filesystem path where the local source is located, e.g.:

```bash
export LOCAL_PROTOS=/path/to/voltha-protos
export LOCAL_LIB_GO=/path/to/voltha-lib-go
```

Then run `make local-protos` and/or `make local-lib-go` as is appropriate to
copy them into the `vendor` directory.

> NOTE: That the files in the `vendor` directory are no longer what is in the
> most recent commit, and it will take manual `git` intervention to put the
> original files back.

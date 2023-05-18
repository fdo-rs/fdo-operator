# Build the manager binary
FROM registry.access.redhat.com/ubi9/go-toolset:1.19 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /opt/app-root/src
# Copy the Go Modules manifests
COPY --chown=1001:0 go.mod go.mod
COPY --chown=1001:0 go.sum go.sum

# Copy the dependencies
COPY --chown=1001:0 vendor/ vendor/

# Copy the go source
COPY --chown=1001:0 main.go main.go
COPY --chown=1001:0 api/ api/
COPY --chown=1001:0 controllers/ controllers/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.2
WORKDIR /
COPY --from=builder /opt/app-root/src/manager .
USER 1001

ENTRYPOINT ["/manager"]

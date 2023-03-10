FROM registry.access.redhat.com/ubi9-minimal:9.1.0 AS builder
 
RUN microdnf install -y tar gzip

# install go 1.19.6
RUN curl -O -J https://dl.google.com/go/go1.19.6.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.19.6.linux-amd64.tar.gz
RUN ln -s /usr/local/go/bin/go /usr/local/bin/go

USER root

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY runners/ runners/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.1.0

WORKDIR /
COPY --from=builder /workspace/manager .

USER 65532:65532

ENTRYPOINT ["/manager"]
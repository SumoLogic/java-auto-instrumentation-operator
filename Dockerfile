# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/manager/main.go main.go
COPY pkg/ pkg/
COPY version/ version/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

FROM gcr.io/distroless/static:nonroot

ENV OPERATOR=/usr/local/bin/java-auto-instrumetation-operator \
    USER_UID=1001 \
    USER_NAME=java-auto-instrumetation-operator

# install operator binary
COPY --from=builder /workspace/manager ${OPERATOR}

USER ${USER_UID}:${USER_UID}

ENTRYPOINT ["/usr/local/bin/java-auto-instrumetation-operator"]

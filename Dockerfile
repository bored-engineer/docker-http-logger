# It is faster to cross-compile if the application can be built without CGO
FROM --platform=${BUILDPLATFORM} cgr.dev/chainguard/go:latest AS builder

# Building from / or a directory in GOPATH can cause problems
WORKDIR /build

# Fetch the Golang dependencies
RUN --mount=type=cache,target=/go/pkg \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download && go mod verify

# Copy only the Golang source code into the builder
COPY *.go /build/

# Cross-compile, using the cached packages and caching the build artifacts
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -o /app .

# The final build layer
FROM cgr.dev/chainguard/busybox:latest
COPY --from=builder /app /app
ENTRYPOINT ["/app"]

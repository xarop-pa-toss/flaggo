# syntax=docker/dockerfile:1
#
# Flaggo — multi-stage image definition.
#
# Three images come out of this one file, selected with --target:
#
#   docker build --target api      -t flaggo/api      .
#   docker build --target frontend -t flaggo/frontend .
#   docker build --target client   -t flaggo/client   .
#
# Why multi-stage: the build needs a full Go toolchain, the runtime does not.
# Keeping the toolchain in a throwaway stage means the final image contains
# only a static binary — nothing to patch, no shell, no shell-injection surface.
#
# Why the build needs no network at runtime: CGO_ENABLED=0 produces a static
# binary that runs on distroless/scratch with no libc.
#
# Why each final stage is ~10 MB instead of ~350 MB: the golang:* builder
# layer is never part of the shipped image.

ARG GO_VERSION=1.27


# ---------------------------------------------------------------- deps stage
# Downloading dependencies is its own layer so it is cached independently of
# your source code. Change go.mod/go.sum -> this layer rebuilds.
# Change a .go file -> this layer is reused. That is the whole point.
FROM golang:${GO_VERSION}-alpine AS deps
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download


# ------------------------------------------------------------- build stages
FROM deps AS build-api
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/flaggo ./cmd/flaggo

FROM deps AS build-frontend
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/flaggo-web ./cmd/flaggo-web

FROM deps AS build-client
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w" \
    -o /out/flaggo-client ./examples/client


# -------------------------------------------------------------- final stages
# distroless/static:nonroot -> no shell, no package manager, no CVEs in a
# libc we don't even use. Ships a nonroot user (uid 65532) and CA certs.
#
# HEALTHCHECK uses the exec form (["/flaggo", "healthcheck"]) so it needs no
# shell — which is exactly why flaggo ships a healthcheck subcommand.
# ENV PORT matters: the healthcheck subcommand reads $PORT to know where to
# probe. Keep the two in sync.
FROM gcr.io/distroless/static:nonroot AS api
COPY --from=build-api /out/flaggo /flaggo
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/flaggo", "healthcheck"]
ENTRYPOINT ["/flaggo"]

FROM gcr.io/distroless/static:nonroot AS frontend
COPY --from=build-frontend /out/flaggo-web /flaggo-web
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/flaggo-web", "healthcheck"]
ENTRYPOINT ["/flaggo-web"]

# The data-plane client is a worker: it polls, it serves nothing.
# No EXPOSE, no HEALTHCHECK, no port — by design.
FROM gcr.io/distroless/static:nonroot AS client
COPY --from=build-client /out/flaggo-client /flaggo-client
USER nonroot:nonroot
ENTRYPOINT ["/flaggo-client"]

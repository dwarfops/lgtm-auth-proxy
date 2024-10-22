FROM golang:1.22-bookworm AS builder

WORKDIR /src

RUN apt-get update \
 && apt-get install -qy ca-certificates \
 && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-w -s -extldflags '-static'" -a .

FROM debian:bookworm-slim

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/lgtm-auth-proxy /usr/local/bin/lgtm-auth-proxy
COPY config.toml /etc/lgtm-auth-proxy/config.toml

RUN useradd -r -u 10001 -g nogroup durin

USER durin

ENTRYPOINT ["/usr/local/bin/lgtm-auth-proxy"]

CMD ["server"]

EXPOSE 8000

LABEL maintainer="Dwarf Ops, Inc. <foss@dwarfops.com>"

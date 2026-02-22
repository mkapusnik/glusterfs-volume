FROM golang:1.20-bullseye AS builder

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      libglusterfs-dev uuid-dev glusterfs-client ca-certificates pkg-config build-essential && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY . .
RUN go build -o /out/docker-volume-glusterfs

FROM scratch AS artifact
COPY --from=builder /out/docker-volume-glusterfs /docker-volume-glusterfs

FROM debian:bullseye-slim AS plugin

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      glusterfs-client libgfapi0 tini ca-certificates && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /var/lib/glusterfs-volume

COPY --from=builder /out/docker-volume-glusterfs /docker-volume-glusterfs
RUN ln -sf /usr/bin/tini /tini

ENTRYPOINT ["/tini", "--"]
CMD ["/docker-volume-glusterfs"]

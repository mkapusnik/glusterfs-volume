FROM golang:1.20-bookworm AS builder

<<<<<<< Updated upstream
RUN yum -y update
RUN yum install -y glusterfs-api

FROM base AS builder

RUN yum install -y glusterfs-api-devel gcc curl
RUN curl -k https://dl.google.com/go/go1.14.1.linux-amd64.tar.gz | tar xz -C /usr/local

ENV PATH=/usr/local/go/bin:$PATH

FROM builder AS build

WORKDIR /src
COPY . .
RUN go build -o /usr/local/bin/docker-volume-glusterfs .

WORKDIR /tini
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini/tini
RUN chmod +x /tini/tini
=======
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      libglusterfs-dev uuid-dev glusterfs-client ca-certificates pkg-config build-essential && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY . .
RUN go build -o /out/docker-volume-glusterfs

FROM debian:bookworm-slim AS plugin
>>>>>>> Stashed changes

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      glusterfs-client libgfapi0 tini ca-certificates && \
    rm -rf /var/lib/apt/lists/*

<<<<<<< Updated upstream
COPY --from=build /tini/tini /usr/local/bin/tini
COPY --from=build /usr/local/bin/docker-volume-glusterfs /usr/local/bin/docker-volume-glusterfs

ENTRYPOINT ["tini", "--"]
CMD ["docker-volume-glusterfs"]

=======
COPY --from=builder /out/docker-volume-glusterfs /docker-volume-glusterfs
RUN ln -sf /usr/bin/tini /tini

ENTRYPOINT ["/tini", "--"]
CMD ["/docker-volume-glusterfs"]
>>>>>>> Stashed changes

FROM gluster/glusterfs-client as base

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

FROM base as plugin

COPY --from=build /tini/tini /usr/local/bin/tini
COPY --from=build /usr/local/bin/docker-volume-glusterfs /usr/local/bin/docker-volume-glusterfs

ENTRYPOINT ["tini", "--"]
CMD ["docker-volume-glusterfs"]


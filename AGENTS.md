# Project Guidelines

## Overview
Docker volume plugin that provisions Docker volumes as subdirectories in a single GlusterFS volume. The plugin runs in a container with FUSE and exposes docker.volumedriver/1.0 over a Unix socket.

## Prerequisites
- Docker 20.10+ with plugin support and BuildKit
- Linux host with FUSE, CAP_SYS_ADMIN, and /dev/fuse available
- GlusterFS servers reachable from the host
- Optional: local test cluster via docker compose (gluster-cluster)

## Repository Structure
- main.go: plugin entrypoint (env, gfapi mount, handler)
- driver.go: Docker VolumeDriver implementation
- gfs.go: gfapi operations and mount/unmount helpers
- Dockerfile, docker-compose.yml: build and runtime
- gluster-cluster/: two-node Gluster test cluster
- config.json: plugin manifest
- Makefile: build, image, plugin, push, test, shell, deploy, clean

## Environment Variables
- GFS_VOLUME: Gluster volfile-id (e.g., gv0)
- GFS_SERVERS: comma-separated list of volfile-server hosts
- LOGFILE: optional path inside plugin for logs

## Commands
- make build: create Docker plugin from local sources
- make image: build base image for plugin filesystem
- make plugin: assemble plugin rootfs and config
- make push plugin=REGISTRY/NAME: push plugin to registry
- make clean: remove plugin artifacts and containers
- make shell: dev shell in builder container
- make test: integration smoke test against local gluster compose target
- gluster-cluster/Makefile: start/clean a 2-node test cluster

Notes:
- Configure via make variables: registry, plugin, context, node, volume, servers, alias
- Override defaults: make test servers=server1,server2 volume=gv0

## Development Workflow
- Build locally: make build
- Configure and enable plugin:
  - docker plugin set $(plugin) GFS_VOLUME=<gv0> GFS_SERVERS=<server1,server2>
  - docker plugin enable $(plugin)
- Create/list/mount docker volumes as usual
- For integration testing: start gluster-cluster, create gv0 per its README, then make test servers=server1,server2 volume=gv0

## Lint/Typecheck (local)
- Format: go fmt ./...
- Vet: go vet ./...
- Typecheck: go build ./...
- Suggested Make targets to add:
  - make lint: go fmt ./... && go vet ./...
  - make typecheck: go build ./...

## Testing
- Unit tests: none yet; recommended to add for driver logic by introducing interfaces to mock:
  - gfapi operations (Init, Mount, Stat, Mkdir, Readdir, Rmdir, Unlink)
  - exec calls for mount/umount
- Integration tests:
  - Start test cluster (gluster-cluster)
  - Create a gv0 volume per gluster-cluster/README
  - Run: make test servers=server1,server2 volume=gv0

## Multi-arch Builds
- Current flow targets linux/amd64. To support x86_64 and arm64:
  - Use docker buildx and multi-arch base images
  - Add a Makefile target for buildx (platforms=linux/amd64,linux/arm64)
  - Verify gluster client base supports both architectures

## Coding Conventions
- Go modules; avoid hard-coded hosts/volumes (use env)
- No logging timestamps (outer runtime provides them)
- Guard shared maps and state with RWMutex
- Do not log secrets or sensitive paths

## Known Issues / TODO
- Update module path from example.com/docker-volume-glusterfs to the actual repository path
- Bump Go version and modernize Dockerfile toolchain
- Add unit tests with mockable interfaces for gfapi and exec
- Add build tags or relocate minio/s3.go (references undefined types, won’t compile)
- Remove environment-specific defaults in Makefile (servers) or require override
- Add buildx multi-arch target and CI workflow
- Add CONTRIBUTING.md and issue templates
- Document scope=global vs scope=local behaviors and cleanup lifecycle

## Troubleshooting
- Docker daemon/plugin logs: check your platform’s Docker logs
- Inspect plugin rootfs with runc under /var/run/docker/plugins/runtime-root/plugins.moby/
- Stale FUSE mounts: unmount and retry; ensure CAP_SYS_ADMIN and /dev/fuse are present

## CI (suggested)
- Lint and typecheck on PRs
- Build plugin image
- Optional: integration job using gluster-cluster; add arm64 via qemu if needed

## Supported Architectures
- x86_64
- arm64 (pending buildx and base image verification)

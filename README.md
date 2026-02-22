# GlusterFS volume plugin

## Overview
Docker volume plugin that provisions Docker volumes of a GlusterFS volume. The plugin runs in a container with FUSE.

## Instalation
    
    docker plugin install --alias gfs ghcr.io/mkapusnik/glusterfs-volume

## Usage

docker-compose.yml:

    volumes:
        gfs:
            driver: gfs
            driver_opts:
                server: glusterfs.server # comma-separated list of volfile-servers hosts
                volume: "volname/subdir" # name of the gfs volume (volfile-id) + optionally path to subdirectory
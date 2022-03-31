#!/bin/bash

IMAGE_NAME=triton-client
VERSION=v1.0
CONTAINER_IMAGE=$IMAGE_NAME:${VERSION}

sudo docker run --rm --add-host=host.docker.internal:host-gateway -d -p8302:8302 $CONTAINER_IMAGE

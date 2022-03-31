#!/bin/bash

IMAGE_NAME=parsec-client
VERSION=v1.0
CONTAINER_IMAGE=$IMAGE_NAME:${VERSION}

sudo docker run --rm -d -p8300:8300 -v /home/parsec/run:/run/parsec $CONTAINER_IMAGE

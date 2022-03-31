#!/bin/bash

IMAGE_NAME=parsec-client
VERSION=v1.0
CONTAINER_IMAGE=$IMAGE_NAME:${VERSION}

sudo docker build . --cache-from $CONTAINER_IMAGE -t $CONTAINER_IMAGE

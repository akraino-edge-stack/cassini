#!/bin/bash

IMAGE_NAME=triton-client
VERSION=v1.0
CONTAINER_IMAGE=$IMAGE_NAME:${VERSION}

# download and get triton client py library
curl -LJO https://github.com/triton-inference-server/server/releases/download/v2.17.0/tritonserver2.17.0-jetpack4.6.tgz
tar -zxf tritonserver2.17.0-jetpack4.6.tgz ./clients/python/tritonclient-2.17.0-py3-none-manylinux2014_aarch64.whl
mv ./clients/python/tritonclient-2.17.0-py3-none-manylinux2014_aarch64.whl ./
rm -Rf clients

# start build docker
sudo docker build . --cache-from $CONTAINER_IMAGE -t $CONTAINER_IMAGE

# clean up download
rm tritonserver2.17.0-jetpack4.6.tgz
rm tritonclient-2.17.0-py3-none-manylinux2014_aarch64.whl

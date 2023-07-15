#!/usr/bin/env bash

VERSION=$(cat VERSION)

# https://www.docker.com/blog/multi-arch-build-and-images-the-simple-way/
PLATFORMS=linux/arm/v7,linux/arm64/v8,linux/amd64

function build_image() {
  docker buildx build         \
     -f Dockerfile.$2         \
     --push                   \
     --platform ${PLATFORMS}  \
     --tag slinkgo/disco:$1 .
}

case $1 in
  alpine)
    build_image "${VERSION}-$1" $1
    build_image "${VERSION}" $1
    build_image "$1" $1
    build_image latest $1
  ;;
  debian)
    build_image "${VERSION}-$1" $1
    build_image "$1" $1
  ;;
  *)
    echo "supported targets: debian, alpine"
  ;;
esac

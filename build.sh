#!/usr/bin/env bash

VERSION=$(cat VERSION)

case $1 in
  debian|alpine)
    docker build -f Dockerfile.$1 --tag slink/disco:${VERSION}-$1 .
  ;;
  *)
    echo "supported targets: debian, alpine"
  ;;
esac

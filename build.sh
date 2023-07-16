#!/usr/bin/env bash

export DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "${DIR}" || exit 1

VERSION_SHORT=$(echo "$(cat ${DIR}/build/VERSION)")
VERSION_LONG=$(echo "v$(cat ${DIR}/build/VERSION) ($(git describe --abbrev=8 --dirty --always))")

# https://www.docker.com/blog/multi-arch-build-and-images-the-simple-way/
PLATFORMS=linux/arm/v7,linux/arm64/v8,linux/amd64

function prepare() {
  cat ${DIR}/build/logo.txt | sed -e "s/VERSION/${VERSION_LONG}/g" > ${DIR}/server/logo.txt
}
function build_image() {
  docker buildx build         \
     -f Dockerfile.$2         \
     --push                   \
     --platform ${PLATFORMS}  \
     --tag slinkgo/disco:$1 .
}


case $1 in
  prepare)
    prepare
  ;;
  alpine)
    prepare
    build_image "${VERSION_SHORT}-$1" $1
    build_image "${VERSION_SHORT}" $1
    build_image "$1" $1
    build_image latest $1
    rm ${DIR}/src/logo.txt > /dev/null
  ;;
  debian)
    prepare
    build_image "${VERSION_SHORT}-$1" $1
    build_image "$1" $1
    rm ${DIR}/src/logo.txt > /dev/null
  ;;
  *)
    echo "supported targets: debian, alpine"
  ;;
esac


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
  FILE=$1
  shift
  TAGS=""
  for var in "$@"
  do
      TAGS="${TAGS} -t slinkgo/disco:$var"
  done
  echo "TAGS: $TAGS"
  docker buildx create --use
  docker buildx build         \
     -f Dockerfile.$FILE      \
     --push                   \
     --platform ${PLATFORMS}  \
     ${TAGS} .
     docker buildx rm
}


case $1 in
  prepare)
    prepare
  ;;
  alpine)
    prepare
    build_image $1 "${VERSION_SHORT}-$1" "${VERSION_SHORT}" "$1" "latest"
    rm ${DIR}/server/logo.txt 2> /dev/null
  ;;
  debian)
    prepare
    build_image $1 "$1" "${VERSION_SHORT}-$1"
    rm ${DIR}/server/logo.txt 2> /dev/null
  ;;
  bin)
    prepare
    templ generate && \
    go build -ldflags "-s -w" -buildmode plugin -o build/inmem.so backend/inmem/registry.go && \
    go build -ldflags="-s -w" -o build/disco ./server
  ;;
  *)
    echo "supported targets: debian, alpine"
  ;;
esac


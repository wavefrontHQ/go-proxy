#!/usr/bin/env bash

which docker
if [[ $? -ne 0 ]] ; then
    echo "Error: docker not found."
    exit 1
fi

CURR_DIR=$(pwd)
VERSION=$1

rm -rf ${CURR_DIR}/build
mkdir -p ${CURR_DIR}/build

PREFIX="goproxy-build"

echo "Building go-proxy docker image..."
docker build -t ${PREFIX}:$VERSION -f ${CURR_DIR}/pkg/Dockerfile-build .

echo "Copying files from docker image..."
CONTAINER_ID=`docker run -d -t ${PREFIX}:${VERSION}`
docker cp ${CONTAINER_ID}:/go/src/github.com/wavefronthq/go-proxy/build/linux  ${CURR_DIR}/build/

echo "Cleaning up build container..."
docker stop ${CONTAINER_ID}
docker container rm ${CONTAINER_ID}

echo "Done. Packages available under: ${CURR_DIR}/build"

#!/usr/bin/env bash

PKG=cmd/wavefront-proxy/proxy.go
PKG_NAME=wavefront-proxy
CURR_DIR=$(pwd)
OUT_DIR=$CURR_DIR/build

VERSION=$1
ALL=$2
TAG=`git describe --exact-match --tags 2>/dev/null`
COMMIT=`git rev-parse --short HEAD`
BRANCH=`git rev-parse --abbrev-ref HEAD`
LDFLAGS="-X main.commit=${COMMIT} -X main.branch=${BRANCH}"

if [[ -n $VERSION ]] ; then
    LDFLAGS="${LDFLAGS} -X main.version=${VERSION}"
fi

if [[ -n $TAG ]] ; then
    LDFLAGS="${LDFLAGS} -X main.tag=${TAG}"
fi

rm -rf ${OUT_DIR}
mkdir ${OUT_DIR}

echo "Output directory: ${OUT_DIR}"
echo "Building ${PKG_NAME} executables"

############################################################
# Build executables

# Default to linux
platforms=("linux/amd64")
if [[ -n $ALL ]] ; then
    platforms=("windows/amd64" "darwin/amd64" "linux/amd64")
fi

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

	output_name=${PKG_NAME}
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    exec_dir=${OUT_DIR}/${GOOS}/${GOARCH}
    mkdir -p ${exec_dir}
    echo "Building ${output_name} for ${GOOS}-${GOARCH}"
    env GOOS=$GOOS GOARCH=$GOARCH go build -o ${exec_dir}/${output_name} -ldflags "${LDFLAGS}" ${PKG}

	if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting build.'
        exit 1
    fi
done


############################################################
# Platform packaging

FPM_PATH=$(which fpm)
if [[ $? -ne 0 ]] ; then
    echo "fpm not found! Cannot build platform packages"
    exit 1
fi

echo "Building ${PKG_NAME} platform packages"

STAGING_DIR=$OUT_DIR/staging
BIN_DIR=$STAGING_DIR/usr/bin

rm -rf $STAGING_DIR
mkdir -p $BIN_DIR

cp $OUT_DIR/linux/amd64/* $BIN_DIR
cp -R pkg/etc $STAGING_DIR
cp -R pkg/usr $STAGING_DIR


packages=("deb" "rpm" "tar")
for pkg in "${packages[@]}"
do
    echo "Building .${pkg} package"

    #TODO: include version
    fpm \
        --after-install pkg/post-install.sh \
        --before-remove pkg/pre-remove.sh \
        --after-remove pkg/post-remove.sh \
        --deb-no-default-config-files \
        --description "Proxy for sending data to Wavefront." \
        --license "Apache 2.0" \
        --maintainer "Wavefront <support@wavefront.com>" \
        --name wavefront-proxy \
        --package $OUT_DIR/linux/amd64 \
        -s dir \
        -t ${pkg} \
        -C $STAGING_DIR \
        etc usr

    if [[ $? -ne 0 ]] ; then
        echo "Error building ${pkg} package."
    fi
done

# Move files from staging into appropriate build directories

# clean up staging dir
rm -rf $STAGING_DIR

echo "Done."

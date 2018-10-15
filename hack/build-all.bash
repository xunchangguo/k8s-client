#!/usr/bin/env bash
# Copyright 2017 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
#
# This script will build dep and calculate hash for each
# (DEP_BUILD_PLATFORMS, DEP_BUILD_ARCHS) pair.
# DEP_BUILD_PLATFORMS="linux" DEP_BUILD_ARCHS="amd64" ./hack/build-all.bash
# can be called to build only for linux-amd64

set -e

DEP_ROOT=$(git rev-parse --show-toplevel)
VERSION=$(git describe --tags --dirty)
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null)
DATE=$(date "+%Y-%m-%d")
BUILD_PLATFORM=$(uname -a | awk '{print tolower($1);}')
IMPORT_DURING_SOLVE=${IMPORT_DURING_SOLVE:-false}

if [[ "$(pwd)" != "${DEP_ROOT}" ]]; then
  echo "you are not in the root of the repo" 1>&2
  echo "please cd to ${DEP_ROOT} before running this script" 1>&2
  exit 1
fi

GO_BUILD_CMD="go build -a"
GO_BUILD_LDFLAGS="-s -w -X main.commitHash=${COMMIT_HASH} -X main.buildDate=${DATE} -X main.version=${VERSION} -X main.flagImportDuringSolve=${IMPORT_DURING_SOLVE}"

if [[ -z "${DEP_BUILD_PLATFORMS}" ]]; then
    DEP_BUILD_PLATFORMS="linux windows"
fi

if [[ -z "${DEP_BUILD_ARCHS}" ]]; then
    DEP_BUILD_ARCHS="amd64 386"
fi

mkdir -p "${DEP_ROOT}/release"

for OS in ${DEP_BUILD_PLATFORMS[@]}; do
  for ARCH in ${DEP_BUILD_ARCHS[@]}; do
    NAME="cli-${OS}-${ARCH}"
    if [[ "${OS}" == "windows" ]]; then
      NAME="${NAME}.exe"
    fi

    CGO_ENABLED=0

    echo "Building for ${OS}/${ARCH} with CGO_ENABLED=${CGO_ENABLED}"
    GOARCH=${ARCH} GOOS=${OS} CGO_ENABLED=${CGO_ENABLED} ${GO_BUILD_CMD} -ldflags "${GO_BUILD_LDFLAGS}"\
     -o "${DEP_ROOT}/release/${NAME}" ./cmd/cli
    shasum -a 256 "${DEP_ROOT}/release/${NAME}" > "${DEP_ROOT}/release/${NAME}".sha256
  done
done
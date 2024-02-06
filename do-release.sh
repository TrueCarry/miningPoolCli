#!/bin/bash

# miningPoolCli – open-source tonuniverse mining pool client

# Copyright (C) 2021 tonuniverse.com

# This file is part of miningPoolCli.

# miningPoolCli is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.

# miningPoolCli is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with miningPoolCli.  If not, see <https://www.gnu.org/licenses/>.

help(){
    echo "Usage: 
    ./do-release.sh {linux|windows|darwin} {amd64|arm64}"
    exit 2
}

CLI_NAME="miningPoolCli"
PLATFORM=$1
ARCH=$2

if [ $PLATFORM = "windows" ]; then
    export GOOS=windows
elif [ $PLATFORM = "linux" ]; then
    export GOOS=linux
elif [ $PLATFORM = "darwin" ]; then
    export GOOS=darwin
else 
    echo "Invalid platform"; help
fi

if [ $ARCH = "amd64" ]; then
    export GOARCH=amd64
elif [ $ARCH = "arm64" ]; then
    export GOARCH=arm64
else
    export GOARCH=amd64
fi

echo "envs set: GOOS=${GOOS} GOARCH=${GOARCH}"

PARSE_VER=$(awk '/BuildVersion/{print $NF}' config/version.go)
BUILD_VERSION=$(echo $PARSE_VER | sed 's/^"\(.*\)"$/\1/')

FOLDER="${CLI_NAME}-${BUILD_VERSION}"

printf "Creating release v${BUILD_VERSION}\n\n"

go build

mkdir $FOLDER
touch "${FOLDER}/VERSION_${BUILD_VERSION}_${GOOS}_${GOARCH}"

cp LICENSE README.md $FOLDER

case $PLATFORM in
  linux)

    cp $CLI_NAME $FOLDER 
    cp hiveos_configs/* $FOLDER
    cp -r miner_blob $FOLDER
    sed -i -e "s/CUSTOM_VERSION=/CUSTOM_VERSION=${BUILD_VERSION}/g" $FOLDER/h-manifest.conf
    tar -zcvf "${CLI_NAME}-${BUILD_VERSION}-${GOARCH}-linux.tar.gz" $FOLDER
    ;;
  windows) 
    cp "${CLI_NAME}.exe" $FOLDER
    cp -r miner_blob $FOLDER
    zip -r "${CLI_NAME}-${BUILD_VERSION}-${GOARCH}-windows.zip" $FOLDER
    ;;
  darwin)
    cp $CLI_NAME $FOLDER
    cp -r miner_blob $FOLDER
    tar -zcvf "${CLI_NAME}-${BUILD_VERSION}-${GOARCH}-darwin.tar.gz" $FOLDER
    ;;
esac

rm -rf $FOLDER
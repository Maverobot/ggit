#!/bin/bash

########################################################################
# Package the binaries built on Travis-CI as an AppImage
# By Simon Peter 2016
# For more information, see http://appimage.org/
########################################################################

set -e

# Returns tag of HEAD or continueous
function get_version() {
	tag=$(git tag -l --points-at HEAD)
	if [ -z "$tag" ]; then
		echo "continuous"
	else
		echo "$tag"
	fi
}

export ARCH="$(arch)"

# TODO: automate this VERSION variable
export VERSION=$(get_version)

APP=ggit
LOWERAPP=${APP,,}

mkdir -p "$HOME/$APP/$APP.AppDir/usr/"

BUILD_PATH="$(pwd)"

cd "$HOME/$APP/"

wget -q https://github.com/probonopd/AppImages/raw/master/functions.sh -O ./functions.sh
. ./functions.sh

cd $APP.AppDir

cp "${BUILD_PATH}/${LOWERAPP}" "AppRun"

########################################################################
# Copy desktop and icon file to AppDir for AppRun to pick them up
########################################################################

cp "${BUILD_PATH}/appimage/${LOWERAPP}.desktop" .
cp "${BUILD_PATH}/appimage/${LOWERAPP}.png" .

########################################################################
# Copy in the dependencies that cannot be assumed to be available
# on all target systems
########################################################################

copy_deps

########################################################################
# Patch away absolute paths; it would be nice if they were relative
########################################################################

find . -type f -exec sed -i -e 's|/usr|././|g' {} \;
find . -type f -exec sed -i -e 's@././/bin/env@/usr/bin/env@g' {} \;

########################################################################
# AppDir complete
# Now packaging it as an AppImage
########################################################################

cd .. # Go out of AppImage

mkdir -p ../out/
generate_type2_appimage

#!/usr/bin/env bash
# Copyright 2020 Comcast Cable Communications Management, LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

########################################################################
# Usage:
#   //go:generate build-plugin debug.so         
#
########################################################################

callPath=$(pwd)
scriptName=$(basename "$0")
scriptDir=$( cd "$(dirname "${BASH_SOURCE}")" ; pwd -P )
pushd $scriptDir/.. >> /dev/null
codeDir=`pwd`
project=$(basename $codeDir)
echo "🚀 ${scriptName} started..."
popd >> /dev/null
echo "-----------------------------"

finish() {
	echo 
	echo "-----------------------------"
	echo "🏁 ${scriptName} finished"
	echo
	exit ${1:0}
}


# If an argument is given, use that as the location for the plugin's code
# If a second argument is given, use that as the output file name


path=${1:-$callPath}
outputFile=${2:-$(basename $callPath).so}
if [[ $outputFile == "main.so" ]]; then
	outputFile="$(basename $(dirname $callPath)).so"
fi

GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_VERSION="$(git describe --tags --abbrev=0)"

echo "path: $path"
echo "output file: $outputFile"
echo "GIT_COMMIT: $GIT_COMMIT"
echo "GIT_VERSION: $GIT_VERSION"


pushd $path >> /dev/null

go build \
	-buildmode=plugin \
	"-ldflags=-X main.GitCommit=$GIT_COMMIT -X main.GitVersion=$GIT_VERSION" \
	-o $outputFile ./

popd >> /dev/null

if [[ ! -f "$path/$outputFile" ]]; then
	echo "error:  could not build plugin"
	return finish 1
fi


finish

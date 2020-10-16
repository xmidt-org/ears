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
#   addlicense.sh           # runs check mode only
#   addlicense.sh apply     # applies license to all *.go and *.sh files
#
########################################################################

scriptName=$(basename "$0")
scriptDir=$( cd "$(dirname "${BASH_SOURCE}")" ; pwd -P )
pushd $scriptDir/.. >> /dev/null
codeDir=`pwd`
project=$(basename $codeDir)
echo "🚀 ${scriptName} started..."
echo "-----------------------------"

finish() {
	echo 
	echo "-----------------------------"
	echo "🏁 ${scriptName} finished"
	echo
	exit ${1:0}
}


cmd=$(which addlicense || true)
if [[ $cmd == "" ]]; then
	cmd="$GOPATH/bin/addlicense"

	if [[ ! -f ${cmd} ]]; then
		echo "* Addlicense doesn't exist.  Installing..."
		go get -u github.com/google/addlicense
		go install github.com/google/addlicense
		echo
	fi

fi


if [[ ! -f ${cmd} ]]; then
	echo "  - No addlicense command found.  Exiting."
	finish 1
fi

echo "* Processing files:"
files=$(find ${codeDir} -type f -name "*.sh" -o -name "*.go" )
if [[ $files == "" ]]; then
	echo "  - No matching files found to process.  Exiting."
	finish 1
fi
echo -e "$files"
echo


echo "* Running addlicense command:"

doCheck="-check"
if [[ $1 != "" ]]; then
	doCheck=""
	echo "  - Applying changes"
else
	echo "  - Running in check mode only"
fi

${cmd} -v \
	-c 'Comcast Cable Communications Management, LLC' \
	-l apache ${doCheck} ${files}

finish

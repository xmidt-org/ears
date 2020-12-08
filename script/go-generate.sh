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
#   go-generate.sh         
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

echo "* Processing directories:"
dirs=$(find ${codeDir}  -type d -not -path '*/\.*' )
if [[ $dirs == "" ]]; then
	echo "  - No matching directories found to process.  Exiting."
	finish 1
fi
echo -e "$dirs"
echo


echo "* Running go generate command:"


p=$(dirname $codeDir)
for d in $dirs
do
	echo "Processing: ${d#"$p"}"
	pushd ${d} >> /dev/null && go generate && popd >> /dev/null
done

finish

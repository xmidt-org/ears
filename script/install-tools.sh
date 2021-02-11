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


tools=(addlicense go-enum moq)

toolCmds=(
	"go get -u github.com/google/addlicense && go install github.com/google/addlicense"
	"go get -u github.com/searKing/golang && go install github.com/searKing/golang/tools/cmd/go-enum"
	"go get -u github.com/matryer/moq && go install github.com/matryer/moq"
)

count=0
for cmd in ${tools[@]}; do
	echo
	echo "-- Checking $cmd"

	c=$(which $cmd || true)

	if [[ $c == "" ]]; then
		echo " * Not found via which.  Checking GOPATH"

		c="$GOPATH/bin/$cmd"
		if [[ ! -f ${c} ]]; then
			echo " * $cmd doesn't exist under GOPATH.  Installing..."
			eval "${toolCmds[$count]}"
      echo " * Install complete."
		else
			echo " * $cmd exists under GOPATH"
    fi

		if [[ ! -f ${c} ]]; then
			echo " * ERROR:  No $cmd command found.  $c"
		else
			echo " * Success: $c"
		fi
  else
		echo " * Success: $c"
	fi

	count=$(( $count + 1 ))

done

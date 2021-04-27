// Copyright 2020 Comcast Cable Communications Management, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hash

import "github.com/xorcare/pointer"

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	HashPath      string `json:"hashPath,omitempty"`
	ResultPath    string `json:"resultPath,omitempty"` // if omitted hashPath will be used as result path
	Metadata      *bool  `json:"metadata,omitempty"`   // if true result will go to metadata, otherwise payload
	HashAlgorithm string `json:"hashAlgorithm,omitempty"`
	Key           string `json:"key,omitempty"` // optional key for certain hash algorithms
	Mod           *int   `json:"mod,omitempty"` // optional modulo for certain integer based hashes
}

var DefaultConfig = Config{
	HashPath:      ".",
	ResultPath:    "",
	Metadata:      pointer.Bool(false),
	HashAlgorithm: "md5",
	Key:           "",
	Mod:           pointer.Int(0),
}

type Filter struct {
	config Config
}

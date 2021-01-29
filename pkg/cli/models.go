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

package cli

type ArgType string

const (
	ArgTypeInt         ArgType = "int"
	ArgTypeIntSlice    ArgType = "[]int"
	ArgTypeFloat       ArgType = "float"
	ArgTypeFloatSlice  ArgType = "[]float"
	ArgTypeString      ArgType = "string"
	ArgTypeStringSlice ArgType = "[]string"
	ArgTypeBool        ArgType = "bool"
	ArgTypeBoolSlice   ArgType = "[]bool"
)

type Argument struct {
	Name        string
	Shorthand   string
	Type        ArgType
	Description string
	Default     interface{}
	LookupKey   string
	Persistent  bool
}

type Config struct {
	EnvPrefix string
	Name      string
	Paths     []string
}

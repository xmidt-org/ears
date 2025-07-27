// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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

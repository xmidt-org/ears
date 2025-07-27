// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package config

type Exporter interface {
	JSONExporter
	YAMLExporter
	Stringer
}

type Importer interface {
	JSONImporter
	YAMLImporter
}

type Stringer interface {
	String() string
}

type YAMLExporter interface {
	YAML() (string, error)
}

type YAMLImporter interface {
	FromYAML(in string) error
}

type JSONExporter interface {
	JSON() (string, error)
}

type JSONImporter interface {
	FromJSON(in string) error
}

type Error struct {
	Err error
}

type InvalidArgumentError struct {
	Err error
}

type DataParseError struct {
	Err error
}

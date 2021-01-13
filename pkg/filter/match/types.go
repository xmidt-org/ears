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

package match

import (
	"context"

	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xorcare/pointer"
)

type Matcher interface {
	Match(ctx context.Context, event event.Event) bool
}

//go:generate rm -f modetype_enum.go
//go:generate go-enum -type=ModeType -linecomment -sql=false
type ModeType int

const (
	ModeUnknown ModeType = iota // unknown
	ModeAllow                   // allow
	ModeDeny                    // deny
)

//go:generate rm -f matchertype_enum.go
//go:generate go-enum -type=MatcherType -linecomment -sql=false
type MatcherType int

const (
	MatcherUnknown MatcherType = iota // unknown
	MatcherRegex                      // regex
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	Mode    ModeType    `json:"mode,omitempty"`
	Matcher MatcherType `json:"matcher,omitempty"`
	Pattern *string     `json:"pattern,omitempty"`
}

var DefaultConfig = Config{
	Mode:    ModeAllow,
	Matcher: MatcherRegex,
	Pattern: pointer.String(`^.*$`),
}

type Filter struct {
	matcher Matcher

	config Config
}

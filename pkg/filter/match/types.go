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
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter/match/comparison"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
	"sync"
)

type Matcher interface {
	Match(event event.Event) bool
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
	MatcherUnknown      MatcherType = iota // unknown
	MatcherRegex                           // regex
	MatcherPattern                         // pattern
	MatcherPatternRegex                    // patternregex
	MatcherComparison                      // comparison
)

// Config can be passed into NewFilter() in order to configure
// the behavior of the sender.
type Config struct {
	Mode            ModeType                       `json:"mode,omitempty"`            // allow or deny
	Matcher         MatcherType                    `json:"matcher,omitempty"`         // regex, pattern, patternregex, comparison
	Pattern         interface{}                    `json:"pattern,omitempty"`         // single pattern
	Patterns        []interface{}                  `json:"patterns,omitempty"`        // list of patterns
	PatternsLogic   string                         `json:"patternsLogic,omitempty"`   // AND, OR only needed for pattern matches with multiple patterns
	Path            string                         `json:"path,omitempty"`            // path to match pattern against, if omitted uses payload as default
	ExactArrayMatch *bool                          `json:"exactArrayMatch,omitempty"` // if true pattern array must match payload array exactly, otherwise the pattern array can be a partial of the payload array
	Comparison      *comparison.Comparison         `json:"comparison,omitempty"`      // equal or notEqual comparison
	ComparisonTree  *comparison.ComparisonTreeNode `json:"comparisonTree,omitempty"`  // boolean tree of comparisons
}

var DefaultConfig = Config{
	Mode:            ModeAllow,
	Matcher:         MatcherRegex,
	Pattern:         `^.*$`,
	ExactArrayMatch: pointer.Bool(true),
}

type Filter struct {
	sync.RWMutex
	matcher                       Matcher
	config                        Config
	name                          string
	plugin                        string
	tid                           tenant.Id
	successCounter                int
	errorCounter                  int
	filterCounter                 int
	successVelocityCounter        int
	errorVelocityCounter          int
	filterVelocityCounter         int
	currentSuccessVelocityCounter int
	currentErrorVelocityCounter   int
	currentFilterVelocityCounter  int
	currentSec                    int64
}

// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package match

import (
	"github.com/xmidt-org/ears/pkg/event"
	"github.com/xmidt-org/ears/pkg/filter"
	"github.com/xmidt-org/ears/pkg/filter/match/comparison"
	"github.com/xmidt-org/ears/pkg/tenant"
	"github.com/xorcare/pointer"
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
	matcher Matcher
	config  Config
	name    string
	plugin  string
	tid     tenant.Id
	filter.MetricFilter
}

// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package comparison

import (
	"reflect"
	"strings"

	"github.com/xmidt-org/ears/pkg/event"
)

type Matcher struct {
	comparisonTree *ComparisonTreeNode
	comparison     *Comparison
	patternsLogic  string
}

type ComparisonTreeNode struct {
	Logic      string                `json:"logic,omitempty"` // and, or
	Comparison *Comparison           `json:"comparison,omitempty"`
	ChildNodes []*ComparisonTreeNode `json:"childNodes,omitempty"`
}

type Comparison struct {
	Equal    []map[string]interface{} `json:"equal,omitempty"`
	NotEqual []map[string]interface{} `json:"notEqual,omitempty"`
	//GreaterThan []map[string]interface{} `json:"greaterThan,omitempty"`
	//LessThan    []map[string]interface{} `json:"lessThan,omitempty"`
}

func NewMatcher(comparisonTree *ComparisonTreeNode, comparison *Comparison, patternsLogic string) (*Matcher, error) {
	return &Matcher{comparisonTree: comparisonTree, comparison: comparison, patternsLogic: patternsLogic}, nil
}

func (m *Matcher) Match(event event.Event) bool {
	if m == nil || event == nil {
		return false
	}
	if m.comparison != nil {
		return m.compare(event, m.comparison, m.patternsLogic)
	} else {
		return m.traverseTree(event, m.comparisonTree)
	}
}

func (m *Matcher) traverseTree(evt event.Event, tree *ComparisonTreeNode) bool {
	if tree == nil {
		return true
	}
	compResult := true
	treeResult := true
	if strings.ToLower(tree.Logic) == "or" {
		treeResult = false
		compResult = false
	}
	for _, childNode := range tree.ChildNodes {
		if strings.ToLower(tree.Logic) == "or" {
			treeResult = treeResult || m.traverseTree(evt, childNode)
		} else {
			treeResult = treeResult && m.traverseTree(evt, childNode)
		}
	}
	if tree.Comparison != nil {
		compResult = m.compare(evt, tree.Comparison, tree.Logic)
	}
	if strings.ToLower(tree.Logic) == "or" {
		if len(tree.ChildNodes) == 0 {
			return compResult
		} else if tree.Comparison == nil {
			return treeResult
		}
		return treeResult || compResult
	} else {
		if len(tree.ChildNodes) == 0 {
			return compResult
		} else if tree.Comparison == nil {
			return treeResult
		}
		return treeResult && compResult
	}
}

func (m *Matcher) compare(evt event.Event, cmp *Comparison, logic string) bool {
	if evt == nil || cmp == nil {
		return true
	}
	andLogic := true
	if strings.ToLower(logic) == "or" {
		andLogic = false
	}
	foundOneEqual := false
	for _, eq := range cmp.Equal {
		for b, a := range eq {
			var aObj, bObj interface{}
			aObj = a
			bObj = b
			switch aT := a.(type) {
			case string:
				if strings.HasPrefix(aT, "{") && strings.HasSuffix(aT, "}") {
					aObj, _, _ = evt.GetPathValue(aT[1 : len(aT)-1])
				}
			}
			if strings.HasPrefix(b, "{") && strings.HasSuffix(b, "}") {
				bObj, _, _ = evt.GetPathValue(b[1 : len(b)-1])
			}
			if reflect.DeepEqual(aObj, bObj) {
				foundOneEqual = true
				if !andLogic {
					return true
				}
			} else {
				if andLogic {
					return false
				}
			}
		}
	}
	for _, eq := range cmp.NotEqual {
		for b, a := range eq {
			var aObj, bObj interface{}
			aObj = a
			bObj = b
			switch aT := a.(type) {
			case string:
				if strings.HasPrefix(aT, "{") && strings.HasSuffix(aT, "}") {
					aObj, _, _ = evt.GetPathValue(aT[1 : len(aT)-1])
				}
			}
			if strings.HasPrefix(b, "{") && strings.HasSuffix(b, "}") {
				bObj, _, _ = evt.GetPathValue(b[1 : len(b)-1])
			}
			if !reflect.DeepEqual(aObj, bObj) {
				foundOneEqual = true
				if !andLogic {
					return true
				}
			} else {
				if andLogic {
					return false
				}
			}
		}
	}
	if foundOneEqual || (len(cmp.Equal) == 0 && len(cmp.NotEqual) == 0) {
		return true
	}
	return false
}

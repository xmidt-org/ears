// Copyright 2021 Comcast Cable Communications Management, LLC
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

package template

import (
	"sync"
	tpl "text/template"

	"github.com/Masterminds/sprig"
)

// --------------------------------------------------

type base struct {
	sync.RWMutex
	optionMissingKey OptionMissingKey
	tpl              *tpl.Template
}

func (b *base) withMissingKeyAction(action OptionMissingKey) error {
	b.Lock()
	defer b.Unlock()
	b.optionMissingKey = action
	return nil
}

// init will make sure we have a template object created.  Will
// immediately return if it's already created.
func (b *base) init() {
	b.Lock()
	defer b.Unlock()

	if b.tpl != nil {
		return
	}

	b.tpl = tpl.New("text").Option("missingkey=" + b.optionMissingKey.String())

	// Needs to be a sprig.TxtFunctMap() -- otherwise, characters like "+" will
	// get HTML escaped
	b.tpl = b.tpl.Funcs(sprig.TxtFuncMap())
}

// ==============================================================
// Interface options
// ==============================================================

type options interface {
	withMissingKeyAction(action OptionMissingKey) error
}

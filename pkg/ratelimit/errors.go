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

package ratelimit

import "github.com/xmidt-org/ears/pkg/errs"

type InvalidUnitError struct {
	BadUnit int
}

func (e *InvalidUnitError) Error() string {
	return errs.String("InvalidUnitError", map[string]interface{}{"unit": e.BadUnit}, nil)
}

type LimitReached struct {
}

func (e *LimitReached) Error() string {
	return errs.String("LimitReached", nil, nil)
}

type BackendError struct {
	Source error
}

func (e *BackendError) Error() string {
	return errs.String("BackendError", nil, e.Source)
}

func (e *BackendError) Unwrap() error {
	return e.Source
}

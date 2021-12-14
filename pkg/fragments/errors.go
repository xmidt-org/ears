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

package fragments

import (
	"github.com/xmidt-org/ears/pkg/errs"
	"github.com/xmidt-org/ears/pkg/tenant"
)

func (e *InvalidFragmentError) Unwrap() error {
	return e.Err
}

func (e *InvalidFragmentError) Error() string {
	return errs.String("InvalidFragmentError", nil, e.Err)
}

type FragmentNotFoundError struct {
	TenantId     tenant.Id
	FragmentName string
}

func (e *FragmentNotFoundError) Error() string {
	return errs.String("FragmentNotFoundError", map[string]interface{}{"fragmentName": e.FragmentName, "orgId": e.TenantId.OrgId, "appId": e.TenantId.AppId}, nil)
}

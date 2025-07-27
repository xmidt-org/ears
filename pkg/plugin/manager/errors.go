// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package manager

import "github.com/xmidt-org/ears/pkg/errs"

func (e *OpenPluginError) Unwrap() error {
	return e.Err
}

func (e *OpenPluginError) Error() string {
	return errs.String("OpenPluginError", nil, e.Err)
}

func (e *NewPluginerError) Unwrap() error {
	return e.Err
}

func (e *NewPluginerError) Error() string {
	return errs.String("NewPluginerError", nil, e.Err)
}

func (e *VariableLookupError) Unwrap() error {
	return e.Err
}

func (e *VariableLookupError) Error() string {
	return errs.String("VariableLookupError", nil, e.Err)
}

func (e *InvalidConfigError) Unwrap() error {
	return e.Err
}

func (e *InvalidConfigError) Error() string {
	return errs.String("InvalidConfigError", nil, e.Err)
}

func (e *NewPluginerNotImplementedError) Unwrap() error {
	return nil
}

func (e *NewPluginerNotImplementedError) Error() string {
	return "NewPluginerNotImplementedError"
}

func (e *NewSendererNotImplementedError) Unwrap() error {
	return nil
}

func (e *NewSendererNotImplementedError) Error() string {
	return "NewSendererNotImplementedError"
}

//

func (e *NewReceivererNotImplementedError) Unwrap() error {
	return nil
}

func (e *NewReceivererNotImplementedError) Error() string {
	return "NewReceivererNotImplementedError"
}

func (e *NewFiltererNotImplementedError) Unwrap() error {
	return nil
}

func (e *NewFiltererNotImplementedError) Error() string {
	return "NewFiltererNotImplementedError"
}

func (e *AlreadyRegisteredError) Unwrap() error {
	return nil
}

func (e *AlreadyRegisteredError) Error() string {
	return "AlreadyRegisteredError"
}

func (e *NotRegisteredError) Unwrap() error {
	return nil
}

func (e *NotRegisteredError) Error() string {
	return "NotRegisteredError"
}

func (e *NotFoundError) Unwrap() error {
	return nil
}

func (e *NotFoundError) Error() string {
	return "NotFoundError"
}

func (e *NilPluginError) Unwrap() error {
	return nil
}

func (e *NilPluginError) Error() string {
	return "NilPluginError"
}

// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package quota

import "github.com/xmidt-org/ears/pkg/errs"

type ConfigNotFoundError struct {
	configKey string
}

func (e *ConfigNotFoundError) Error() string {
	return errs.String("ConfigNotFoundError", map[string]interface{}{"configKey": e.configKey}, nil)
}

type BadConfigError struct {
	configKey   string
	configValue string
}

func (e *BadConfigError) Error() string {
	return errs.String("BadConfigError", map[string]interface{}{"configKey": e.configKey, "configValue": e.configValue}, nil)
}

type NoEarsInstances struct {
}

func (e *NoEarsInstances) Error() string {
	return errs.String("NoEarsInstances", nil, nil)
}

// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package tenantstorerfx

type UnsupportedTenantStorageError struct {
	storageType string
}

func (e *UnsupportedTenantStorageError) Error() string {
	return "UnsupportedStorageError: (storageType=" + e.storageType + ")"
}

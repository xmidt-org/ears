// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package routestorerfx

type UnsupportedRouteStorageError struct {
	storageType string
}

func (e *UnsupportedRouteStorageError) Error() string {
	return "UnsupportedStorageError: (storageType=" + e.storageType + ")"
}

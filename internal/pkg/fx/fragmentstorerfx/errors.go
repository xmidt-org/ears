// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package fragmentstorerfx

type UnsupportedFragmentStorageError struct {
	storageType string
}

func (e *UnsupportedFragmentStorageError) Error() string {
	return "UnsupportedStorageError: (storageType=" + e.storageType + ")"
}

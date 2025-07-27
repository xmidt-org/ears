// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package nodestatemanagerfx

type UnsupportedNodeStateStorageError struct {
	storageType string
}

func (e *UnsupportedNodeStateStorageError) Error() string {
	return "UnsupportedStorageError: (storageType=" + e.storageType + ")"
}

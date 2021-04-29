package tenantstorerfx

type UnsupportedTenantStorageError struct {
	storageType string
}

func (e *UnsupportedTenantStorageError) Error() string {
	return "UnsupportedStorageError: (storageType=" + e.storageType + ")"
}

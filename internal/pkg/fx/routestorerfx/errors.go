package routestorerfx

type UnsupportedStorageError struct {
	storageType string
}

func (e *UnsupportedStorageError) Error() string {
	return "UnsupportedStorageError: (storageType=" + e.storageType + ")"
}

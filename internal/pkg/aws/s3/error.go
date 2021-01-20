package s3

// https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully

// AWS S3 will produce the following errors
// https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html#ErrorCodeList
//
// Of this list, the following three will be marked as "Temporary"
//
// * RequestTimeout
// * ServiceUnavailable
// * SlowDown

// ==============================================================
// Error Types
// ==============================================================

// Error is the basic s3 error object interface.
type Error interface {
	error
	String() string
	Cause() error
}

// ParameterError allows type switching to more easily know if
// Key() and Value() can be called on the error object
type ParameterError interface {
	Error
	Key() string
	Value() interface{}
}

// RequestError allows type switching to more easily know if
// Temporary() and Url() can be called on the error object.  For now,
// Temporary() will be set to true if AWS comes back with a
// "RequestTimeout", "ServiceUnavailable", or "SlowDown" coded error.
type RequestError interface {
	Error
	Temporary() bool
	Url() string
}

// ==============================================================
// Classification
//	- To drive resulting behavior
// ==============================================================

// IsTemporary returns true if err is temporary.
func IsTemporary(err error) bool {
	type temporary interface {
		Temporary() bool
	}

	te, ok := err.(temporary)
	return ok && te.Temporary()
}

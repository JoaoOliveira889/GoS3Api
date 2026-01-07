package upload

import "errors"

var (
	ErrBucketNameRequired  = errors.New("bucket name is required")
	ErrFileNotFound        = errors.New("file not found in storage")
	ErrInvalidFileType     = errors.New("file type not allowed or malicious content detected")
	ErrBucketAlreadyExists = errors.New("bucket already exists")
	ErrOperationTimeout    = errors.New("the operation timed out")
)

package operations

import (
	"capyfile/files"
)

const ErrorCodeS3FileUploadFailure = "FILE_S3_FILE_UPLOAD_FAILURE"

func NewS3FileUploadFailureError(origErr error) *S3FileUploadFailureError {
	return &S3FileUploadFailureError{
		Data: &S3FileUploadFailureErrorData{
			OrigErr: origErr,
		},
	}
}

type S3FileUploadFailureError struct {
	files.FileProcessingError
	Data *S3FileUploadFailureErrorData
}

type S3FileUploadFailureErrorData struct {
	OrigErr error
}

func (e *S3FileUploadFailureError) Code() string {
	return ErrorCodeS3FileUploadFailure
}

func (e *S3FileUploadFailureError) Error() string {
	return "failed to upload file to S3"
}

const ErrorCodeS3FileUrlCanNotBeRetrieved = "FILE_S3_FILE_URL_CAN_NOT_BE_RETRIEVED"

func NewS3FileUrlCanNotBeRetrievedError(origErr error) *S3FileUrlCanNotBeRetrievedError {
	return &S3FileUrlCanNotBeRetrievedError{
		Data: &S3FileUrlCanNotBeRetrievedErrorData{
			OrigErr: origErr,
		},
	}
}

type S3FileUrlCanNotBeRetrievedError struct {
	files.FileProcessingError
	Data *S3FileUrlCanNotBeRetrievedErrorData
}

type S3FileUrlCanNotBeRetrievedErrorData struct {
	OrigErr error
}

func (e *S3FileUrlCanNotBeRetrievedError) Code() string {
	return ErrorCodeS3FileUrlCanNotBeRetrieved
}

func (e *S3FileUrlCanNotBeRetrievedError) Error() string {
	return "file cannot be uploaded to S3"
}

package operations

import (
	"capyfile/files"
)

const ErrorCodeFileMimeTypeIsNotAllowed = "FILE_MIME_TYPE_IS_NOT_ALLOWED"

func NewFileMimeTypeIsNotAllowedError(allowedMimeTypes []string, givenMimeType string) *FileMimeTypeIsNotAllowedError {
	return &FileMimeTypeIsNotAllowedError{
		Data: &FileMimeTypeIsNotAllowedErrorData{
			AllowedMimeTypes: allowedMimeTypes,
			GivenMimeType:    givenMimeType,
		},
	}
}

type FileMimeTypeIsNotAllowedError struct {
	files.FileProcessingError

	Data *FileMimeTypeIsNotAllowedErrorData
}

type FileMimeTypeIsNotAllowedErrorData struct {
	AllowedMimeTypes []string
	GivenMimeType    string
}

func (e *FileMimeTypeIsNotAllowedError) Code() string {
	return ErrorCodeFileMimeTypeIsNotAllowed
}

func (e *FileMimeTypeIsNotAllowedError) Error() string {
	return "file MIME type is not allowed"
}

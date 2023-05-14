package operations

import (
	"capyfile/files"
)

const ErrorCodeFileSizeIsTooBig = "FILE_SIZE_IS_TOO_BIG"

func NewFileSizeIsTooBigError(maxFileSize, givenFileSize int64) *FileSizeIsTooBigError {
	return &FileSizeIsTooBigError{
		Data: &FileSizeIsTooBigErrorData{
			MaxFileSize:   maxFileSize,
			GivenFileSize: givenFileSize,
		},
	}
}

type FileSizeIsTooBigError struct {
	files.FileProcessingError

	Data *FileSizeIsTooBigErrorData
}

type FileSizeIsTooBigErrorData struct {
	MaxFileSize   int64
	GivenFileSize int64
}

func (e *FileSizeIsTooBigError) Code() string {
	return ErrorCodeFileSizeIsTooBig
}

func (e *FileSizeIsTooBigError) Error() string {
	return "file size is too big"
}

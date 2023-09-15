package operations

import (
	"capyfile/files"
)

const ErrorCodeFileInputIsUnwritable = "FILE_INPUT_IS_UNWRITABLE"

func NewFileInputIsUnwritableError(originalError error) *FileInputIsUnwritableError {
	return &FileInputIsUnwritableError{
		Data: &FileInputIsUnwritableErrorData{
			OriginalError: originalError,
		},
	}
}

type FileInputIsUnwritableError struct {
	files.FileProcessingError

	Data *FileInputIsUnwritableErrorData
}

type FileInputIsUnwritableErrorData struct {
	OriginalError error
}

func (e *FileInputIsUnwritableError) Code() string {
	return ErrorCodeFileInputIsUnwritable
}

func (e *FileInputIsUnwritableError) Error() string {
	return "file input is unwritable"
}

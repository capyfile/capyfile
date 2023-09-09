package operations

import (
	"capyfile/files"
)

const ErrorCodeFileInputIsUnreadable = "FILE_INPUT_IS_UNREADABLE"

func NewFileInputIsUnreadableError(originalError error) *FileInputIsUnreadableError {
	return &FileInputIsUnreadableError{
		Data: &FileInputIsUnreadableErrorData{
			OriginalError: originalError,
		},
	}
}

type FileInputIsUnreadableError struct {
	files.FileProcessingError

	Data *FileInputIsUnreadableErrorData
}

type FileInputIsUnreadableErrorData struct {
	OriginalError error
}

func (e *FileInputIsUnreadableError) Code() string {
	return ErrorCodeFileInputIsUnreadable
}

func (e *FileInputIsUnreadableError) Error() string {
	return "file input is unreadable"
}

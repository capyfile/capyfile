package operations

import "capyfile/files"

// The file processing errors that can be shared between the operations. For example,
// it can be unreadable/unwritable file. These errors are more generic, and without
// many details they can be shown to the client

const ErrorCodeFileIsUnreadable = "FILE_IS_UNREADABLE"

func NewFileIsUnreadableError(origErr error) *FileIsUnreadableError {
	return &FileIsUnreadableError{
		Data: &FileIsUnreadableErrorData{
			OrigErr: origErr,
		},
	}
}

type FileIsUnreadableError struct {
	files.FileProcessingError

	Data *FileIsUnreadableErrorData
}

type FileIsUnreadableErrorData struct {
	OrigErr error
}

func (e *FileIsUnreadableError) Code() string {
	return ErrorCodeFileIsUnreadable
}

func (e *FileIsUnreadableError) Error() string {
	return "file cannot be read"
}

const ErrorCodeFileIsUnwritable = "FILE_IS_UNWRITABLE"

func NewFileIsUnwritableError(origErr error) *FileIsUnwritableError {
	return &FileIsUnwritableError{
		Data: &FileIsUnwritableErrorData{
			OrigErr: origErr,
		},
	}
}

type FileIsUnwritableError struct {
	files.FileProcessingError

	Data *FileIsUnwritableErrorData
}

type FileIsUnwritableErrorData struct {
	OrigErr error
}

func (e *FileIsUnwritableError) Code() string {
	return ErrorCodeFileIsUnwritable
}

func (e *FileIsUnwritableError) Error() string {
	return "file cannot be written"
}

const ErrorCodeFileInfoCanNotBeRetrieved = "FILE_INFO_CAN_NOT_BE_RETRIEVED"

func NewFileInfoCanNotBeRetrievedError(origErr error) *FileSizeInfoCanNotBeRetrievedError {
	return &FileSizeInfoCanNotBeRetrievedError{
		Data: &FileSizeInfoCanNotBeRetrievedErrorData{
			OrigErr: origErr,
		},
	}
}

type FileSizeInfoCanNotBeRetrievedError struct {
	files.FileProcessingError

	Data *FileSizeInfoCanNotBeRetrievedErrorData
}

type FileSizeInfoCanNotBeRetrievedErrorData struct {
	OrigErr error
}

func (e *FileSizeInfoCanNotBeRetrievedError) Code() string {
	return ErrorCodeFileInfoCanNotBeRetrieved
}

func (e *FileSizeInfoCanNotBeRetrievedError) Error() string {
	return "file info can not be retrieved"
}

const ErrorCodeFileReadOffsetCanNotBeSet = "FILE_READ_OFFSET_CAN_NOT_BE_SET"

func NewFileReadOffsetCanNotBeSetError(origErr error) *FileReadOffsetCanNotBeSetError {
	return &FileReadOffsetCanNotBeSetError{
		Data: &FileReadOffsetCanNotBeSetErrorData{
			OrigErr: origErr,
		},
	}
}

type FileReadOffsetCanNotBeSetError struct {
	files.FileProcessingError

	Data *FileReadOffsetCanNotBeSetErrorData
}

type FileReadOffsetCanNotBeSetErrorData struct {
	OrigErr error
}

func (e *FileReadOffsetCanNotBeSetError) Code() string {
	return ErrorCodeFileReadOffsetCanNotBeSet
}

func (e *FileReadOffsetCanNotBeSetError) Error() string {
	return "file read offset can not be set"
}
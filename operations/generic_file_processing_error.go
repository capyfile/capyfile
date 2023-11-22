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

const ErrorCodeFileCanNotBeOpened = "FILE_CAN_NOT_BE_OPENED"

func NewFileCanNotBeOpenedError(origErr error) *FileCanNotBeOpenedError {
	return &FileCanNotBeOpenedError{
		Data: &FileCanNotBeOpenedErrorData{
			OrigErr: origErr,
		},
	}
}

type FileCanNotBeOpenedError struct {
	files.FileProcessingError

	Data *FileCanNotBeOpenedErrorData
}

type FileCanNotBeOpenedErrorData struct {
	OrigErr error
}

func (e *FileCanNotBeOpenedError) Code() string {
	return ErrorCodeFileCanNotBeOpened
}

func (e *FileCanNotBeOpenedError) Error() string {
	return "file can not be opened"
}

const ErrorCodeFileCanNotBeClosed = "FILE_CAN_NOT_BE_CLOSED"

func NewFileCanNotBeClosedError(origErr error) *FileCanNotBeClosedError {
	return &FileCanNotBeClosedError{
		Data: &FileCanNotBeClosedErrorData{
			OrigErr: origErr,
		},
	}
}

type FileCanNotBeClosedError struct {
	files.FileProcessingError

	Data *FileCanNotBeClosedErrorData
}

type FileCanNotBeClosedErrorData struct {
	OrigErr error
}

func (e *FileCanNotBeClosedError) Code() string {
	return ErrorCodeFileCanNotBeClosed
}

func (e *FileCanNotBeClosedError) Error() string {
	return "file can not be closed"
}

const ErrorCodeTmpFileCanNotBeCreated = "TMP_FILE_CAN_NOT_BE_CREATED"

func NewTmpFileCanNotBeCreatedError(origErr error) *TmpFileCanNotBeCreatedError {
	return &TmpFileCanNotBeCreatedError{
		Data: &TmpFileCanNotBeCreatedErrorData{
			OrigErr: origErr,
		},
	}
}

type TmpFileCanNotBeCreatedError struct {
	files.FileProcessingError

	Data *TmpFileCanNotBeCreatedErrorData
}

type TmpFileCanNotBeCreatedErrorData struct {
	OrigErr error
}

func (e *TmpFileCanNotBeCreatedError) Code() string {
	return ErrorCodeTmpFileCanNotBeCreated
}

func (e *TmpFileCanNotBeCreatedError) Error() string {
	return "tmp file can not be created"
}

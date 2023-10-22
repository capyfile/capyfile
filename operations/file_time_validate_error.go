package operations

import (
	"capyfile/files"
	"time"
)

const ErrorCodeFileAtimeIsTooOld = "FILE_ATIME_IS_TOO_OLD"

func NewFileAtimeIsTooOldError(minFileAtime time.Time, givenFileAtime time.Time) *FileAtimeIsTooOldError {
	return &FileAtimeIsTooOldError{
		Data: &FileTimeErrorData{
			WantedFileTime: minFileAtime,
			GivenFileTime:  givenFileAtime,
		},
	}
}

type FileAtimeIsTooOldError struct {
	files.FileProcessingError

	Data *FileTimeErrorData
}

type FileTimeErrorData struct {
	WantedFileTime time.Time
	GivenFileTime  time.Time
}

func (e *FileAtimeIsTooOldError) Code() string {
	return ErrorCodeFileAtimeIsTooOld
}

func (e *FileAtimeIsTooOldError) Error() string {
	return "file Atime is too old"
}

const ErrorCodeFileAtimeIsTooNew = "FILE_ATIME_IS_TOO_NEW"

func NewFileAtimeIsTooNewError(maxFileAtime time.Time, givenFileAtime time.Time) *FileAtimeIsTooNewError {
	return &FileAtimeIsTooNewError{
		Data: &FileTimeErrorData{
			WantedFileTime: maxFileAtime,
			GivenFileTime:  givenFileAtime,
		},
	}
}

type FileAtimeIsTooNewError struct {
	files.FileProcessingError

	Data *FileTimeErrorData
}

func (e *FileAtimeIsTooNewError) Code() string {
	return ErrorCodeFileAtimeIsTooNew
}

func (e *FileAtimeIsTooNewError) Error() string {
	return "file Atime is too new"
}

const ErrorCodeFileMtimeIsTooOld = "FILE_MTIME_IS_TOO_OLD"

func NewFileMtimeIsTooOldError(minFileMtime time.Time, givenFileMtime time.Time) *FileMtimeIsTooOldError {
	return &FileMtimeIsTooOldError{
		Data: &FileTimeErrorData{
			WantedFileTime: minFileMtime,
			GivenFileTime:  givenFileMtime,
		},
	}
}

type FileMtimeIsTooOldError struct {
	files.FileProcessingError

	Data *FileTimeErrorData
}

func (e *FileMtimeIsTooOldError) Code() string {
	return ErrorCodeFileMtimeIsTooOld
}

func (e *FileMtimeIsTooOldError) Error() string {
	return "file Mtime is too old"
}

const ErrorCodeFileMtimeIsTooNew = "FILE_MTIME_IS_TOO_NEW"

func NewFileMtimeIsTooNewError(maxFileMtime time.Time, givenFileMtime time.Time) *FileMtimeIsTooNewError {
	return &FileMtimeIsTooNewError{
		Data: &FileTimeErrorData{
			WantedFileTime: maxFileMtime,
			GivenFileTime:  givenFileMtime,
		},
	}
}

type FileMtimeIsTooNewError struct {
	files.FileProcessingError

	Data *FileTimeErrorData
}

func (e *FileMtimeIsTooNewError) Code() string {
	return ErrorCodeFileMtimeIsTooNew
}

func (e *FileMtimeIsTooNewError) Error() string {
	return "file Mtime is too new"
}

const ErrorCodeFileCtimeIsTooOld = "FILE_CTIME_IS_TOO_OLD"

func NewFileCtimeIsTooOldError(minFileCtime time.Time, givenFileCtime time.Time) *FileCtimeIsTooOldError {
	return &FileCtimeIsTooOldError{
		Data: &FileTimeErrorData{
			WantedFileTime: minFileCtime,
			GivenFileTime:  givenFileCtime,
		},
	}
}

type FileCtimeIsTooOldError struct {
	files.FileProcessingError

	Data *FileTimeErrorData
}

func (e *FileCtimeIsTooOldError) Code() string {
	return ErrorCodeFileCtimeIsTooOld
}

func (e *FileCtimeIsTooOldError) Error() string {
	return "file Ctime is too old"
}

const ErrorCodeFileCtimeIsTooNew = "FILE_CTIME_IS_TOO_NEW"

func NewFileCtimeIsTooNewError(maxFileCtime time.Time, givenFileCtime time.Time) *FileCtimeIsTooNewError {
	return &FileCtimeIsTooNewError{
		Data: &FileTimeErrorData{
			WantedFileTime: maxFileCtime,
			GivenFileTime:  givenFileCtime,
		},
	}
}

type FileCtimeIsTooNewError struct {
	files.FileProcessingError

	Data *FileTimeErrorData
}

func (e *FileCtimeIsTooNewError) Code() string {
	return ErrorCodeFileCtimeIsTooNew
}

func (e *FileCtimeIsTooNewError) Error() string {
	return "file Ctime is too new"
}

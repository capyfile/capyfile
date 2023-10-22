package operations

import (
	"capyfile/files"
)

const ErrorCodeFileMetadataCanNotBeWritten = "FILE_METADATA_CAN_NOT_BE_WRITTEN"

func NewFileMetadataCanNotBeWrittenError(origErr error) *FileMetadataCanNotBeWrittenError {
	return &FileMetadataCanNotBeWrittenError{
		Data: &FileMetadataCanNotBeWrittenErrorData{
			OrigErr: origErr,
		},
	}
}

type FileMetadataCanNotBeWrittenError struct {
	files.FileProcessingError
	Data *FileMetadataCanNotBeWrittenErrorData
}

type FileMetadataCanNotBeWrittenErrorData struct {
	OrigErr error
}

func (e *FileMetadataCanNotBeWrittenError) Code() string {
	return ErrorCodeFileMetadataCanNotBeWritten
}

func (e *FileMetadataCanNotBeWrittenError) Error() string {
	return "file metadata can not be written"
}

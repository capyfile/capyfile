package operations

import (
	"capyfile/files"
)

const ErrorCodeFileMetadataCanNotBeWritten = "FILE_METADATA_CAN_NOT_BE_WRITTEN"

func NewFileMetadataCanNotBeWrittenError() *FileMetadataCanNotBeWrittenError {
	return &FileMetadataCanNotBeWrittenError{}
}

type FileMetadataCanNotBeWrittenError struct {
	files.FileProcessingError
}

func (e *FileMetadataCanNotBeWrittenError) Code() string {
	return ErrorCodeFileMetadataCanNotBeWritten
}

func (e *FileMetadataCanNotBeWrittenError) Error() string {
	return "file metadata can not be written"
}

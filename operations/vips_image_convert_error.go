package operations

import (
	"capyfile/files"
)

const ErrorCodeBimgImageProcessorError = "BIMG_IMAGE_PROCESSOR_ERROR"

func NewBimgImageProcessorError(originalError error) *BimgImageProcessorError {
	return &BimgImageProcessorError{
		Data: &BimgImageProcessorErrorData{
			OriginalError: originalError,
		},
	}
}

type BimgImageProcessorError struct {
	files.FileProcessingError

	Data *BimgImageProcessorErrorData
}

type BimgImageProcessorErrorData struct {
	OriginalError error
}

func (e *BimgImageProcessorError) Code() string {
	return ErrorCodeBimgImageProcessorError
}

func (e *BimgImageProcessorError) Error() string {
	return "bimg failed to process the image"
}

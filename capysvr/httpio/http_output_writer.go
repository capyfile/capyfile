package httpio

import (
	"capyfile/capysvc/common"
	"capyfile/files"
	"capyfile/operations"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"golang.org/x/exp/slog"
	"net/http"
)

// Right now this is quire messy, but this is going to be a place
// for output encoding, formatting, and localization.

type ResponseDTO struct {
	Status  string                   `json:"status"`
	Code    string                   `json:"code"`
	Message string                   `json:"message"`
	Files   []FileDTO                `json:"files"`
	Errors  []FileProcessingErrorDTO `json:"errors"`
	Meta    MetaDTO                  `json:"meta"`
}

type FileDTO struct {
	Url              string  `json:"url"`
	Filename         string  `json:"filename"`
	OriginalFilename *string `json:"originalFilename"`

	Mime string `json:"mime"`
	Size int64  `json:"size"`

	Original *OriginalFileDTO `json:"original,omitempty"`

	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type OriginalFileDTO struct {
	Url      *string `json:"url"`
	Filename *string `json:"filename"`

	Mime string `json:"mime"`
	Size int64  `json:"size"`
}

type FileProcessingErrorDTO struct {
	OriginalFilename string `json:"originalFilename"`

	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type MetaDTO struct {
	TotalUploads      int `json:"totalUploads"`
	SuccessfulUploads int `json:"successfulUploads"`
	FailedUploads     int `json:"failedUploads"`
}

// WriteOutput Writes processor's output to the http response.
func WriteOutput(out []files.ProcessableFile, w http.ResponseWriter) error {
	var responseDTO ResponseDTO

	for _, processableFile := range out {
		responseDTO.writeProcessedFile(&processableFile)
	}
	responseDTO.writeStatusAndMeta()

	responseWriterErr := json.NewEncoder(w).Encode(responseDTO)
	if responseWriterErr != nil {
		common.Logger.Error(
			"failed to write output for http",
			slog.Any("error", responseWriterErr),
		)

		return responseWriterErr
	}

	return nil
}

func (dto *ResponseDTO) writeProcessedFile(processableFile *files.ProcessableFile) {
	if !processableFile.HasFileProcessingError() {
		var fileURL string
		if val, ok := processableFile.OperationMetadata[operations.MetadataKeyS3UploadV2FileUrl]; ok {
			fileURL = val.(string)
		}

		originalFilename := processableFile.OriginalFilename()

		mime, _ := processableFile.Mime()
		fileInfo, _ := processableFile.File.Stat()

		//original := &OriginalFileDTO{}

		dto.Files = append(dto.Files, FileDTO{
			Status:  "SUCCESS",
			Code:    "FILE_SUCCESSFULLY_UPLOADED",
			Message: "file successfully uploaded",

			Url:              fileURL,
			Filename:         processableFile.GeneratedFilename(),
			OriginalFilename: &originalFilename,

			Mime: mime.String(),
			Size: fileInfo.Size(),

			//Original: original,
		})

		return
	}

	var errorMessage = ""

	var fileSizeIsTooSmall *operations.FileSizeIsTooSmallError
	var fileSizeIsTooBig *operations.FileSizeIsTooBigError
	var fileMimeTypeIsNotAllowed *operations.FileMimeTypeIsNotAllowedError
	switch {
	case errors.As(processableFile.FileProcessingError, &fileSizeIsTooSmall):
		errorMessage = fmt.Sprintf(
			"file size can not be less than %s",
			humanize.IBytes(uint64(fileSizeIsTooSmall.Data.MinFileSize)),
		)
		break
	case errors.As(processableFile.FileProcessingError, &fileSizeIsTooBig):
		errorMessage = fmt.Sprintf(
			"file size can not be greater than %s",
			humanize.IBytes(uint64(fileSizeIsTooBig.Data.MaxFileSize)),
		)
		break
	case errors.As(processableFile.FileProcessingError, &fileMimeTypeIsNotAllowed):
		errorMessage = fmt.Sprintf(
			"file MIME type \"%s\" is not allowed",
			fileMimeTypeIsNotAllowed.Data.GivenMimeType,
		)
		break
	}

	dto.Errors = append(dto.Errors, FileProcessingErrorDTO{
		Code:             processableFile.FileProcessingError.Code(),
		Status:           "ERROR",
		Message:          errorMessage,
		OriginalFilename: processableFile.OriginalFilename(),
	})
}

func (dto *ResponseDTO) writeStatusAndMeta() {
	successfulUploads := len(dto.Files)
	failedUploads := len(dto.Errors)
	totalUploads := successfulUploads + failedUploads

	var status = "UNKNOWN"
	var code = "NO_FILES_PROVIDED"
	var message = "no files provided"
	if successfulUploads != 0 && failedUploads != 0 {
		status = "PARTIAL"
		code = "PARTIAL"
		message = fmt.Sprintf("successfully uploaded %d of %d files", successfulUploads, totalUploads)
	} else if successfulUploads != 0 {
		status = "SUCCESS"
		code = "SUCCESS"
		message = fmt.Sprintf("successfully uploaded %d file(s)", successfulUploads)
	} else if failedUploads != 0 {
		status = "ERROR"
		code = "ERROR"
		message = fmt.Sprintf("failed to upload %d file(s)", failedUploads)
	}

	dto.Status = status
	dto.Code = code
	dto.Message = message

	dto.Meta = MetaDTO{
		SuccessfulUploads: successfulUploads,
		FailedUploads:     failedUploads,
		TotalUploads:      totalUploads,
	}
}

type ErrorDTO struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteError Writes the error to the http response.
func WriteError(err error, w http.ResponseWriter) error {
	var errorDTO ErrorDTO
	errorDTO.Status = "ERROR"

	var httpErr *HTTPAwareType
	switch {
	case errors.As(err, &httpErr):
		w.WriteHeader(httpErr.HTTPStatusCode())

		errorDTO.Code = httpErr.Code()
		errorDTO.Message = httpErr.Message()

		break
	default:
		w.WriteHeader(http.StatusInternalServerError)

		errorDTO.Code = "UNKNOWN"
		errorDTO.Message = "UNKNOWN"
	}

	responseWriterErr := json.NewEncoder(w).Encode(errorDTO)
	if responseWriterErr != nil {
		common.Logger.Error(
			"failed to write error for http",
			slog.Any("error", responseWriterErr),
		)

		return responseWriterErr
	}

	return nil
}

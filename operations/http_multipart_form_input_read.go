package operations

import (
	"capyfile/capyutils"
	"capyfile/files"
	"net/http"
)

// HttpMultipartFormInputReadOperation reads input from the http request for further processing.
type HttpMultipartFormInputReadOperation struct {
	Name   string
	Params *HttpMultipartFormInputReadOperationParams
	Req    *http.Request
}

type HttpMultipartFormInputReadOperationParams struct {
}

func (o *HttpMultipartFormInputReadOperation) OperationName() string {
	return o.Name
}

func (o *HttpMultipartFormInputReadOperation) AllowConcurrency() bool {
	return false
}

func (o *HttpMultipartFormInputReadOperation) Handle(
	in []files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (out []files.ProcessableFile, err error) {
	multipartFormErr := o.Req.ParseMultipartForm(0)
	// There's `ErrNotMultipart` error type.
	if multipartFormErr != nil {
		if errorCh != nil {
			errorCh <- o.errorBuilder().Error(multipartFormErr)
		}

		return out, multipartFormErr
	}

	for _, fileHeaders := range o.Req.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			uploadedFile, fileOpenErr := fileHeader.Open()
			if fileOpenErr != nil {
				if errorCh != nil {
					errorCh <- o.errorBuilder().Error(fileOpenErr)
				}

				// Perhaps it makes sense to try to open other files.
				continue
			}

			file, fileWriteErr := capyutils.WriteReaderToAppTmpDirectory(uploadedFile)
			if fileWriteErr != nil {
				if errorCh != nil {
					errorCh <- o.errorBuilder().Error(fileWriteErr)
				}

				continue
			}

			pf := files.NewProcessableFile(file.Name())
			pf.Metadata.OriginalFilename = fileHeader.Filename

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Finished("multipart form file read finished", &pf)
			}

			out = append(out, pf)
		}
	}

	return out, nil
}

func (o *HttpMultipartFormInputReadOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *HttpMultipartFormInputReadOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}

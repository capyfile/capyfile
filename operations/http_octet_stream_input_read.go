package operations

import (
	"capyfile/capyutils"
	"capyfile/files"
	"net/http"
)

// HttpOctetStreamInputReadOperation reads input from the http request body for further processing.
type HttpOctetStreamInputReadOperation struct {
	Name   string
	Params *HttpOctetStreamInputReadOperationParams
	Req    *http.Request
}

type HttpOctetStreamInputReadOperationParams struct {
}

func (o *HttpOctetStreamInputReadOperation) OperationName() string {
	return o.Name
}

func (o *HttpOctetStreamInputReadOperation) AllowConcurrency() bool {
	return false
}

func (o *HttpOctetStreamInputReadOperation) Handle(
	in []files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (out []files.ProcessableFile, err error) {
	if o.Req.Body == http.NoBody {
		return out, nil
	}

	f, fileWriteErr := capyutils.WriteReaderToTempFileAndLeaveOpen(o.Req.Body)
	if fileWriteErr != nil {
		if errorCh != nil {
			errorCh <- o.errorBuilder().Error(fileWriteErr)
		}

		return out, fileWriteErr
	}

	pf := files.NewProcessableFile(f)

	if notificationCh != nil {
		notificationCh <- o.notificationBuilder().Finished("octet-stream file read finished", &pf)
	}

	out = append(out, pf)

	return out, nil
}

func (o *HttpOctetStreamInputReadOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *HttpOctetStreamInputReadOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}

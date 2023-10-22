package operations

import (
	"capyfile/files"
	"sync"
)

type FileTypeValidateOperation struct {
	Name   string
	Params *FileTypeValidateOperationParams
}

type FileTypeValidateOperationParams struct {
	AllowedMimeTypes []string
}

func (o *FileTypeValidateOperation) OperationName() string {
	return o.Name
}

func (o *FileTypeValidateOperation) AllowConcurrency() bool {
	return true
}

func (o *FileTypeValidateOperation) Handle(
	in []files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (out []files.ProcessableFile, err error) {
	var wg sync.WaitGroup

	outHolder := newOutputHolder()

	for i := range in {
		wg.Add(1)

		pf := &in[i]

		go func(pf *files.ProcessableFile) {
			defer wg.Done()

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Started("file type validation started", pf)
			}

			if len(o.Params.AllowedMimeTypes) > 0 {
				mime, mimeErr := pf.Mime()
				if mimeErr != nil {
					pf.SetFileProcessingError(
						NewFileMimeTypeCanNotBeDeterminedError(mimeErr),
					)

					if errorCh != nil {
						errorCh <- o.errorBuilder().ProcessableFileError(pf, mimeErr)
					}
					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Failed(
							"file type validation failed with error", pf, mimeErr)
					}

					outHolder.AppendToOut(pf)

					return
				}

				var allowed = false
				for _, allowedMime := range o.Params.AllowedMimeTypes {
					if mime.Is(allowedMime) {
						allowed = true
					}
				}
				if !allowed {
					pf.SetFileProcessingError(
						NewFileMimeTypeIsNotAllowedError(o.Params.AllowedMimeTypes, mime.String()),
					)

					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Finished(
							"file type validation finished with file processing error", pf)
					}

					outHolder.AppendToOut(pf)

					return
				}
			}

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Finished("file type validation finished", pf)
			}

			outHolder.AppendToOut(pf)
		}(pf)
	}

	wg.Wait()

	return outHolder.Out, nil
}

func (o *FileTypeValidateOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *FileTypeValidateOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}

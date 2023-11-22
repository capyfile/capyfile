package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"github.com/spf13/afero"
	"sync"
)

type FileSizeValidateOperation struct {
	Name   string
	Params *FileSizeValidateOperationParams
}

type FileSizeValidateOperationParams struct {
	MinFileSize int64
	MaxFileSize int64
}

func (o *FileSizeValidateOperation) OperationName() string {
	return o.Name
}

func (o *FileSizeValidateOperation) AllowConcurrency() bool {
	return true
}

func (o *FileSizeValidateOperation) Handle(
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
				notificationCh <- o.notificationBuilder().Started("file size validation started", pf)
			}

			file, fileOpenErr := capyfs.Filesystem.Open(pf.Name())
			if fileOpenErr != nil {
				pf.SetFileProcessingError(
					NewFileCanNotBeOpenedError(fileOpenErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, fileOpenErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"file can not be opened", pf, fileOpenErr)
				}

				outHolder.AppendToOut(pf)

				return
			}
			defer func(file afero.File) {
				closeErr := file.Close()
				if closeErr != nil {
					if errorCh != nil {
						errorCh <- o.errorBuilder().ProcessableFileError(pf, fileOpenErr)
					}
					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Failed(
							"file can not be opened", pf, fileOpenErr)
					}
				}
			}(file)

			fileStat, statErr := file.Stat()
			if statErr != nil {
				// This may be related to the specific file, so it makes sense to add a file processing
				// error to the processable file. We can also return more specific error here, but
				// it's not necessary at the moment.
				pf.SetFileProcessingError(
					NewFileInfoCanNotBeRetrievedError(statErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, statErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"file info can not be retrieved", pf, statErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			if o.Params.MinFileSize > 0 {
				if fileStat.Size() < o.Params.MinFileSize {
					pf.SetFileProcessingError(
						NewFileSizeIsTooSmallError(o.Params.MinFileSize, fileStat.Size()),
					)

					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Finished(
							"file size is too small", pf)
					}

					outHolder.AppendToOut(pf)

					return
				}
			}

			if o.Params.MaxFileSize > 0 {
				if fileStat.Size() > o.Params.MaxFileSize {
					pf.SetFileProcessingError(
						NewFileSizeIsTooBigError(o.Params.MaxFileSize, fileStat.Size()),
					)

					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Finished(
							"file size is too big", pf)
					}

					outHolder.AppendToOut(pf)

					return
				}
			}

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Finished("file size is valid", pf)
			}

			outHolder.AppendToOut(pf)
		}(pf)
	}

	wg.Wait()

	return outHolder.Out, nil
}

func (o *FileSizeValidateOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *FileSizeValidateOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}

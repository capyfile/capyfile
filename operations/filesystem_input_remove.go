package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"sync"
)

// FilesystemInputRemoveOperation removes input from the filesystem.
type FilesystemInputRemoveOperation struct {
	Name   string
	Params *FilesystemInputRemoveOperationParams
}

type FilesystemInputRemoveOperationParams struct {
}

func (o *FilesystemInputRemoveOperation) OperationName() string {
	return o.Name
}

func (o *FilesystemInputRemoveOperation) AllowConcurrency() bool {
	return true
}

func (o *FilesystemInputRemoveOperation) Handle(
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
				notificationCh <- o.notificationBuilder().Started("file remove started", pf)
			}

			closeErr := pf.File.Close()
			if closeErr != nil {
				pf.SetFileProcessingError(
					NewFileInputIsUnwritableError(closeErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, closeErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed("file close failed with the error", pf, closeErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			removeErr := capyfs.Filesystem.Remove(pf.File.Name())
			if removeErr != nil {
				pf.SetFileProcessingError(
					NewFileInputIsUnwritableError(removeErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, removeErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed("file remove failed with the error", pf, removeErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Finished("file remove finished", pf)
			}
		}(pf)
	}

	wg.Wait()

	return outHolder.Out, nil
}

func (o *FilesystemInputRemoveOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *FilesystemInputRemoveOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}

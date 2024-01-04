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
	// If set to true, the original file will be removed along with the current
	// processable file.
	RemoveOriginalFile bool
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

			removeErr := capyfs.Filesystem.Remove(pf.Name())
			if removeErr != nil {
				pf.SetFileProcessingError(
					NewFileInputIsUnwritableError(removeErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, removeErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"file remove failed with the error", pf, removeErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			if o.Params.RemoveOriginalFile && pf.OriginalProcessableFile != nil {
				origRemoveErr := capyfs.Filesystem.Remove(pf.OriginalProcessableFile.Name())
				if origRemoveErr != nil {
					pf.SetFileProcessingError(
						NewFileInputIsUnwritableError(origRemoveErr),
					)

					if errorCh != nil {
						errorCh <- o.errorBuilder().ProcessableFileError(pf, origRemoveErr)
					}
					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Failed(
							"original file remove failed with the error", pf, origRemoveErr)
					}

					outHolder.AppendToOut(pf)

					return
				}
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

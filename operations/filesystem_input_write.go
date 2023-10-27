package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"path/filepath"
	"sync"
)

// FilesystemInputWriteOperation writes input to the filesystem.
type FilesystemInputWriteOperation struct {
	Name   string
	Params *FilesystemInputWriteOperationParams
}

type FilesystemInputWriteOperationParams struct {
	// Destination is the target directory to write to.
	Destination string
	// UseOriginalFilename indicates whether to use the original filename or the generated one (nanoid).
	UseOriginalFilename bool
}

func (o *FilesystemInputWriteOperation) OperationName() string {
	return o.Name
}

func (o *FilesystemInputWriteOperation) AllowConcurrency() bool {
	return true
}

func (o *FilesystemInputWriteOperation) Handle(
	in []files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (out []files.ProcessableFile, err error) {
	var wg sync.WaitGroup

	for i := range in {
		wg.Add(1)

		pf := &in[i]

		go func(pf *files.ProcessableFile) {
			defer wg.Done()

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Started("file write started", pf)
			}

			var base string
			if o.Params.UseOriginalFilename {
				base = filepath.Base(pf.OriginalFilename())
			} else {
				base = pf.GeneratedFilename()
			}
			destFilename := filepath.Join(o.Params.Destination, base)

			writeErr := capyfs.FilesystemUtils.WriteReader(destFilename, pf.File)
			if writeErr != nil {
				pf.SetFileProcessingError(
					NewFileInputIsUnwritableError(writeErr),
				)

				if errorCh != nil {
					errorCh <- o.operationError(writeErr, pf)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed("file write failed with error", pf, writeErr)
				}

				return
			}

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Finished("file write finished", pf)
			}
		}(pf)
	}

	wg.Wait()

	return in, nil
}

func (o *FilesystemInputWriteOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *FilesystemInputWriteOperation) operationError(err error, pf *files.ProcessableFile) OperationError {
	return OperationError{
		OperationName: o.Name,
		In:            []files.ProcessableFile{*pf},
		Err:           err,
	}
}
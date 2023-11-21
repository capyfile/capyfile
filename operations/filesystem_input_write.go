package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"github.com/spf13/afero"
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
				// In addition to it, we need to ensure that the extension is relevant.
				// This is for the cases we transform the file to another format, etc.
				ext := filepath.Ext(pf.OriginalFilename())
				if ext != "" {
					base = base[:len(base)-len(ext)] + filepath.Ext(pf.GeneratedFilename())
				}
			} else {
				base = pf.GeneratedFilename()
			}
			destFilename := filepath.Join(o.Params.Destination, base)

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

			writeErr := capyfs.FilesystemUtils.WriteReader(destFilename, file)
			if writeErr != nil {
				pf.SetFileProcessingError(
					NewFileInputIsUnwritableError(writeErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, writeErr)
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

func (o *FilesystemInputWriteOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}

package operations

import (
	"capyfile/capyerr"
	"capyfile/capyutils"
	"capyfile/files"
	"errors"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"os/exec"
	"path/filepath"
	"sync"
)

const ErrorCodeExiftoolMetadataCleanupOperationConfiguration = "EXIFTOOL_METADATA_CLEANUP_OPERATION_CONFIGURATION"

type ExiftoolMetadataCleanupOperation struct {
	Name   string
	Params *ExiftoolMetadataCleanupOperationParams
}

func (o *ExiftoolMetadataCleanupOperation) OperationName() string {
	return o.Name
}

func (o *ExiftoolMetadataCleanupOperation) AllowConcurrency() bool {
	return true
}

type ExiftoolMetadataCleanupOperationParams struct {
	OverwriteOriginalFile bool
}

func (o *ExiftoolMetadataCleanupOperation) Handle(
	in []files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (out []files.ProcessableFile, err error) {
	// First it makes sense to check if the exiftool is installed.
	_, exiftoolVerErr := exec.Command("exiftool", "-ver").Output()
	if exiftoolVerErr != nil {
		if errorCh != nil {
			errorCh <- o.errorBuilder().Error(
				errors.New("exiftool is not installed"),
			)
		}

		return out, capyerr.NewOperationConfigurationError(
			ErrorCodeExiftoolMetadataCleanupOperationConfiguration,
			"exiftool is not installed",
			exiftoolVerErr,
		)
	}

	outHolder := newOutputHolder()

	var wg sync.WaitGroup

	for i := range in {
		wg.Add(1)

		pf := &in[i]

		go func(pf *files.ProcessableFile) {
			defer wg.Done()

			// If this is not empty, then the original file was not overwritten,
			// and now we assign it with the processable file.
			var tmpFilename string

			var args []string
			if o.Params.OverwriteOriginalFile {
				args = []string{"-all:all=", "-overwrite_original", pf.Name()}
			} else {
				tmpDir, tmpDirErr := capyutils.GetAppTmpDirectory()
				if tmpDirErr != nil {
					pf.SetFileProcessingError(
						NewTmpFileCanNotBeCreatedError(tmpDirErr),
					)
					outHolder.AppendToOut(pf)

					if errorCh != nil {
						errorCh <- o.errorBuilder().ProcessableFileError(pf, tmpDirErr)
					}
					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Failed(
							"exiftool failed to write the file metadata", pf, tmpDirErr)
					}

					return
				}
				tmpFilename = filepath.Join(tmpDir, gonanoid.Must())

				args = []string{
					"-all:all=",
					"-o", tmpFilename,
					pf.Name(),
				}
			}
			_, exiftoolErr := exec.Command("exiftool", args...).Output()
			if exiftoolErr != nil {
				pf.SetFileProcessingError(
					NewFileMetadataCanNotBeWrittenError(exiftoolErr),
				)
				outHolder.AppendToOut(pf)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, exiftoolErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"exiftool failed to write the file metadata", pf, exiftoolErr)
				}

				return
			}

			if tmpFilename != "" {
				pf.ReplaceFile(tmpFilename)
			}

			outHolder.AppendToOut(pf)

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Finished("file metadata cleanup has finished", pf)
			}
		}(pf)
	}

	wg.Wait()

	return outHolder.Out, nil
}

func (o *ExiftoolMetadataCleanupOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *ExiftoolMetadataCleanupOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}

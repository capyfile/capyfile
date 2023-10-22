package operations

import (
	"capyfile/capyerr"
	"capyfile/files"
	"errors"
	"os/exec"
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

	var wg sync.WaitGroup

	for i := range in {
		wg.Add(1)

		pf := &in[i]

		go func(pf *files.ProcessableFile) {
			defer wg.Done()

			// Right now this is limited to os filesystem.
			args := []string{"-all:all=", "-overwrite_original", pf.File.Name()}
			_, exiftoolErr := exec.Command("exiftool", args...).Output()
			if exiftoolErr != nil {
				pf.SetFileProcessingError(
					NewFileMetadataCanNotBeWrittenError(exiftoolErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, exiftoolErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"exiftool failed to write the file metadata", pf, exiftoolErr)
				}

				return
			}

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Finished("file metadata cleanup has finished", pf)
			}
		}(pf)
	}

	wg.Wait()

	return in, nil
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

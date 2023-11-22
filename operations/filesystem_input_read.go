package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"github.com/spf13/afero"
)

// FilesystemInputReadOperation reads input from the filesystem for further processing.
type FilesystemInputReadOperation struct {
	Name   string
	Params *FilesystemInputReadOperationParams
}

type FilesystemInputReadOperationParams struct {
	// Target is the target file or directory to read from. Can be a glob pattern.
	Target string
}

func (o *FilesystemInputReadOperation) OperationName() string {
	return o.Name
}

func (o *FilesystemInputReadOperation) AllowConcurrency() bool {
	return false
}

func (o *FilesystemInputReadOperation) Handle(
	in []files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (out []files.ProcessableFile, err error) {
	matches, matchesErr := afero.Glob(capyfs.Filesystem, o.Params.Target)
	if matchesErr != nil {
		if errorCh != nil {
			errorCh <- o.errorBuilder().Error(matchesErr)
		}

		return out, matchesErr
	}

	for _, match := range matches {
		file, fileOpenErr := capyfs.Filesystem.Open(match)
		if fileOpenErr != nil {
			if errorCh != nil {
				errorCh <- o.errorBuilder().Error(fileOpenErr)
			}

			continue
		}

		pf := files.NewProcessableFile(file.Name())

		if notificationCh != nil {
			notificationCh <- o.notificationBuilder().Finished("file read finished", &pf)
		}

		out = append(out, pf)
	}

	return out, nil
}

func (o *FilesystemInputReadOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *FilesystemInputReadOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}

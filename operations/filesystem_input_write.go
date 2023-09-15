package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"path/filepath"
)

// FilesystemInputWriteOperation writes input to the filesystem.
type FilesystemInputWriteOperation struct {
	Params *FilesystemInputWriteOperationParams
}

type FilesystemInputWriteOperationParams struct {
	// Destination is the target directory to write to.
	Destination string
	// UseOriginalFilename indicates whether to use the original filename or the generated one (nanoid).
	UseOriginalFilename bool
}

func (o *FilesystemInputWriteOperation) Handle(in []files.ProcessableFile) ([]files.ProcessableFile, error) {
	for i := range in {
		processableFile := &in[i]

		if processableFile.HasFileProcessingError() {
			continue
		}

		var base string
		if o.Params.UseOriginalFilename {
			base = filepath.Base(processableFile.OriginalFilename())
		} else {
			base = processableFile.GeneratedFilename()
		}
		destFilename := filepath.Join(o.Params.Destination, base)

		writeErr := capyfs.FilesystemUtils.WriteReader(destFilename, processableFile.File)
		if writeErr != nil {
			processableFile.SetFileProcessingError(
				NewFileInputIsUnwritableError(writeErr),
			)
		}
	}

	return in, nil
}

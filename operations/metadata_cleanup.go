package operations

import (
	"capyfile/files"
	"os/exec"
)

type MetadataCleanupOperation struct {
	Params *MetadataCleanupOperationParams
}

type MetadataCleanupOperationParams struct {
}

func (o *MetadataCleanupOperation) Handle(in []files.ProcessableFile) ([]files.ProcessableFile, error) {
	for i := range in {
		processableFile := &in[i]

		if processableFile.HasFileProcessingError() {
			continue
		}

		// Right now this is limited to os filesystem.
		args := []string{"-all:all=", "-overwrite_original", processableFile.File.Name()}
		_, exiftoolErr := exec.Command("exiftool", args...).Output()
		if exiftoolErr != nil {
			processableFile.SetFileProcessingError(
				NewFileMetadataCanNotBeWrittenError(),
			)
		}
	}

	return in, nil
}

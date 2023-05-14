package operations

import (
	"capyfile/files"
)

type FileTypeValidateOperation struct {
	Params *FileTypeValidateOperationParams
}

type FileTypeValidateOperationParams struct {
	AllowedMimeTypes []string
}

func (o *FileTypeValidateOperation) Handle(in []files.ProcessableFile) ([]files.ProcessableFile, error) {
	for i := range in {
		processableFile := &in[i]

		if processableFile.HasFileProcessingError() {
			continue
		}

		if len(o.Params.AllowedMimeTypes) > 0 {
			mime, err := processableFile.Mime()
			if err != nil {
				return in, err
			}

			var allowed = false
			for _, allowedMime := range o.Params.AllowedMimeTypes {
				if mime.Is(allowedMime) {
					allowed = true
				}
			}
			if !allowed {
				processableFile.SetFileProcessingError(
					NewFileMimeTypeIsNotAllowedError(o.Params.AllowedMimeTypes, mime.String()),
				)

				continue
			}
		}
	}

	return in, nil
}

package operations

import (
	"capyfile/files"
)

type FileSizeValidateOperation struct {
	Params *FileSizeValidateOperationParams
}

type FileSizeValidateOperationParams struct {
	MinFileSize int64
	MaxFileSize int64
}

func (o *FileSizeValidateOperation) Handle(in []files.ProcessableFile) ([]files.ProcessableFile, error) {
	for i := range in {
		processableFile := &in[i]

		if processableFile.HasFileProcessingError() {
			continue
		}

		fileStat, err := processableFile.File.Stat()
		if err != nil {
			return in, err
		}

		if o.Params.MinFileSize > 0 {
			if fileStat.Size() < o.Params.MinFileSize {
				processableFile.SetFileProcessingError(
					NewFileSizeIsTooSmallError(o.Params.MinFileSize, fileStat.Size()),
				)

				continue
			}
		}

		if o.Params.MaxFileSize > 0 {
			if fileStat.Size() > o.Params.MaxFileSize {
				processableFile.SetFileProcessingError(
					NewFileSizeIsTooBigError(o.Params.MaxFileSize, fileStat.Size()),
				)

				continue
			}
		}
	}

	return in, nil
}

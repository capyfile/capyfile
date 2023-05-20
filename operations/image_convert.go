package operations

import (
	"capyfile/capyerr"
	"capyfile/capyutils"
	"capyfile/files"
	"fmt"
	"github.com/h2non/bimg"
	"github.com/spf13/afero"
)

const ErrorCodeImageConvertOperationConfiguration = "IMAGE_CONVERT_OPERATION_CONFIGURATION"

var imageConvertAllowedMimeTypes = map[string]bimg.ImageType{
	"image/jpeg":      bimg.JPEG,
	"image/png":       bimg.PNG,
	"image/gif":       bimg.GIF,
	"image/webp":      bimg.WEBP,
	"image/heic":      bimg.HEIF,
	"image/heif":      bimg.HEIF,
	"application/pdf": bimg.PDF,
}

type ImageConvertOperation struct {
	Params *ImageConvertOperationParams
}

type ImageConvertOperationParams struct {
	ToMimeType string
	Quality    string
}

func (p *ImageConvertOperationParams) bimgNumericQuality() int {
	switch p.Quality {
	case "best":
		return 100
	case "high":
		return 75
	case "medium":
		return 50
	case "low":
		return 25
	default:
		return 75
	}
}

func (o *ImageConvertOperation) Handle(in []files.ProcessableFile) (out []files.ProcessableFile, err error) {
	var imageType = bimg.UNKNOWN
	if allowedImageType, ok := imageConvertAllowedMimeTypes[o.Params.ToMimeType]; ok {
		imageType = allowedImageType
	}

	if imageType == bimg.UNKNOWN {
		// This can happen only if there's some misconfiguration.
		return in, capyerr.NewOperationConfigurationError(
			ErrorCodeImageConvertOperationConfiguration,
			fmt.Sprintf("image conversion to \"%s\" MIME type is not supported", o.Params.ToMimeType),
			err,
		)
	}

	for i := range in {
		processableFile := &in[i]

		if processableFile.HasFileProcessingError() {
			continue
		}

		mime, err := processableFile.Mime()
		if err != nil {
			return in, err
		}

		if mime.Is(o.Params.ToMimeType) {
			continue
		}

		fileStat, err := processableFile.File.Stat()
		if err != nil {
			return in, err
		}

		oldImg := make([]byte, fileStat.Size())
		_, err = processableFile.File.ReadAt(oldImg, 0)
		if err != nil {
			return in, err
		}

		var newImg []byte
		newImg, err = bimg.NewImage(oldImg).Process(
			bimg.Options{
				Type:    imageType,
				Quality: o.Params.bimgNumericQuality(),
			})
		if err != nil {
			return in, err
		}

		var newFile afero.File
		newFile, err = capyutils.WriteBytesToTempFileAndLeaveOpen(newImg)
		if err != nil {
			return in, err
		}

		processableFile.ReplaceFile(newFile)
	}

	return in, nil
}

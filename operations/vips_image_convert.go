package operations

import (
	"capyfile/capyerr"
	"capyfile/capyutils"
	"capyfile/files"
	"fmt"
	"github.com/h2non/bimg"
	"sync"
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
	Name   string
	Params *ImageConvertOperationParams
}

func (o *ImageConvertOperation) OperationName() string {
	return o.Name
}

func (o *ImageConvertOperation) AllowConcurrency() bool {
	return true
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

func (o *ImageConvertOperation) Handle(
	in []files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (out []files.ProcessableFile, err error) {
	var imageType = bimg.UNKNOWN
	if allowedImageType, ok := imageConvertAllowedMimeTypes[o.Params.ToMimeType]; ok {
		imageType = allowedImageType
	}

	if imageType == bimg.UNKNOWN {
		if errorCh != nil {
			errorCh <- o.errorBuilder().Error(
				fmt.Errorf(
					"operation misconfiguration: image conversion to \"%s\" MIME type is not supported",
					o.Params.ToMimeType,
				),
			)
		}

		// This can happen only if there's some misconfiguration.
		// We return an error here, because the operation is unusable no matter what input we provide.
		return out, capyerr.NewOperationConfigurationError(
			ErrorCodeImageConvertOperationConfiguration,
			fmt.Sprintf("image conversion to \"%s\" MIME type is not supported", o.Params.ToMimeType),
			err,
		)
	}

	var wg sync.WaitGroup

	outHolder := newOutputHolder()

	for i := range in {
		wg.Add(1)

		pf := &in[i]

		go func(pf *files.ProcessableFile) {
			defer wg.Done()

			mime, mimeErr := pf.Mime()
			if mimeErr != nil {
				pf.SetFileProcessingError(
					NewFileMimeTypeCanNotBeDeterminedError(mimeErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, mimeErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed("can not to determine the file MIME type", pf, mimeErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			if mime.Is(o.Params.ToMimeType) {
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Skipped("file already has wanted MIME type", pf)
				}

				outHolder.AppendToOut(pf)

				return
			}

			//fileStat, statErr := pf.File.Stat()
			//if statErr != nil {
			//	pf.SetFileProcessingError(
			//		NewFileInfoCanNotBeRetrievedError(statErr),
			//	)
			//
			//	if errorCh != nil {
			//		errorCh <- o.errorBuilder().ProcessableFileError(pf, statErr)
			//	}
			//	if notificationCh != nil {
			//		notificationCh <- o.notificationBuilder().Failed("can not retrieve the file info", pf, statErr)
			//	}
			//
			//	outHolder.AppendToOut(pf)
			//
			//	return
			//}
			//
			//oldImg := make([]byte, fileStat.Size())
			//_, readErr := pf.File.ReadAt(oldImg, 0)
			//if readErr != nil {
			//	pf.SetFileProcessingError(
			//		NewFileIsUnreadableError(readErr),
			//	)
			//
			//	if errorCh != nil {
			//		errorCh <- o.errorBuilder().ProcessableFileError(pf, readErr)
			//	}
			//	if notificationCh != nil {
			//		notificationCh <- o.notificationBuilder().Failed("can not read the file", pf, readErr)
			//	}
			//
			//	outHolder.AppendToOut(pf)
			//
			//	return
			//}
			//
			oldImage, oldImageReadErr := bimg.Read(pf.Name())
			if oldImageReadErr != nil {
				pf.SetFileProcessingError(
					NewFileIsUnreadableError(oldImageReadErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, oldImageReadErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed("can not read the file", pf, oldImageReadErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			newImg, imageProcessErr := bimg.NewImage(oldImage).Process(
				bimg.Options{
					Type:    imageType,
					Quality: o.Params.bimgNumericQuality(),
				})
			if imageProcessErr != nil {
				pf.SetFileProcessingError(
					NewBimgImageProcessorError(imageProcessErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, imageProcessErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"bimg failed to process the image transformation", pf, imageProcessErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			newFile, writeErr := capyutils.WriteBytesToAppTmpDirectory(newImg)
			if writeErr != nil {
				pf.SetFileProcessingError(
					NewFileIsUnwritableError(writeErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, writeErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed("can not write the file", pf, writeErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Finished("image conversion has finished", pf)
			}

			pf.ReplaceFile(newFile.Name())

			outHolder.AppendToOut(pf)
		}(pf)
	}

	wg.Wait()

	return outHolder.Out, nil
}

func (o *ImageConvertOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *ImageConvertOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}

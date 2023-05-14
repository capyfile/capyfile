package httpio

import (
	"capyfile/capysvc/common"
	"capyfile/capyutils"
	"capyfile/files"
	"github.com/spf13/afero"
	"golang.org/x/exp/slog"
	"net/http"
)

// ReadInput Reads the input from the request and returns a slice of ProcessableFile.
// It supports both multipart form data and stream.
func ReadInput(r *http.Request) (in []files.ProcessableFile, err error) {
	multipartFormErr := r.ParseMultipartForm(0)
	// There's `ErrNotMultipart` error type.
	if multipartFormErr != nil {
		// If this is not a multipart form data, treat it as a stream
		common.Logger.Debug("reading request body content")

		if r.Body == http.NoBody {
			return in, nil
		}

		var tempFile afero.File
		tempFile, err = capyutils.WriteReaderToTempFileAndLeaveOpen(r.Body)
		if err != nil {
			return in, err
		}

		common.Logger.Info(
			"request body content extracted",
			slog.String("filename", tempFile.Name()),
		)

		return append(in, *files.NewProcessableFile(tempFile)), nil
	}

	common.Logger.Debug("reading multipart form data")

	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			uploadedFile, fileOpenErr := fileHeader.Open()
			if fileOpenErr != nil {
				common.Logger.Info(
					"failed to open uploaded file",
					slog.Any("error", fileOpenErr),
				)

				return in, fileOpenErr
			}

			f, fileWriteErr := capyutils.WriteReaderToTempFileAndLeaveOpen(uploadedFile)
			if fileWriteErr != nil {
				common.Logger.Info(
					"failed to write uploaded file to tmp file",
					slog.Any("error", fileWriteErr),
				)

				return in, fileWriteErr
			}

			common.Logger.Info(
				"multipart form data extracted",
				slog.String("filename", f.Name()),
			)

			processableFile := files.NewProcessableFile(f)
			processableFile.Metadata = &files.ProcessableFileMetadata{
				OriginalFilename: fileHeader.Filename,
			}

			in = append(in, *processableFile)
		}
	}

	return in, nil
}

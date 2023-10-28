package operations

import (
	"capyfile/files"
	"capyfile/operations/filetime"
	"sync"
	"time"
)

type FileTimeValidateOperation struct {
	Name   string
	Params *FileTimeValidateOperationParams

	TimeStatProvider filetime.TimeStatProvider
}

type FileTimeValidateOperationParams struct {
	MinAtime time.Time
	MaxAtime time.Time
	MinMtime time.Time
	MaxMtime time.Time
	MinCtime time.Time
	MaxCtime time.Time
}

func (o *FileTimeValidateOperation) OperationName() string {
	return o.Name
}

func (o *FileTimeValidateOperation) AllowConcurrency() bool {
	return true
}

func (o *FileTimeValidateOperation) Handle(
	in []files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (out []files.ProcessableFile, err error) {
	if o.TimeStatProvider == nil {
		initErr := o.initTimeStatProvider()
		if initErr != nil {
			return nil, initErr
		}
	}

	var wg sync.WaitGroup

	outHolder := newOutputHolder()

	for i := range in {
		wg.Add(1)

		pf := &in[i]

		go func(pf *files.ProcessableFile) {
			defer wg.Done()

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Started("file time validation started", pf)
			}

			fileInfo, statErr := pf.File.Stat()
			if statErr != nil {
				// This may be related to the specific file, so it makes sense to add a file processing
				// error to the processable file. We can also return more specific error here, but
				// it's not necessary at the moment.
				pf.SetFileProcessingError(
					NewFileInfoCanNotBeRetrievedError(statErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, statErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"can not get file info", pf, statErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			timeStat, timeStatErr := o.TimeStatProvider.TimeStat(fileInfo)
			if timeStatErr != nil {
				// This may be related to the specific file, so it makes sense to add a file processing
				// error to the processable file. We can also return more specific error here, but
				// it's not necessary at the moment.
				pf.SetFileProcessingError(
					NewFileInfoCanNotBeRetrievedError(timeStatErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().ProcessableFileError(pf, timeStatErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"can not get file time stat", pf, timeStatErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			if !o.Params.MinAtime.IsZero() {
				if timeStat.Atime.Before(o.Params.MinAtime) {
					pf.SetFileProcessingError(
						NewFileAtimeIsTooOldError(o.Params.MinAtime, fileInfo.ModTime()),
					)

					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Finished(
							"file atime is too old", pf)
					}

					outHolder.AppendToOut(pf)

					return
				}
			}

			if !o.Params.MaxAtime.IsZero() {
				if timeStat.Atime.After(o.Params.MaxAtime) {
					pf.SetFileProcessingError(
						NewFileAtimeIsTooNewError(o.Params.MaxAtime, fileInfo.ModTime()),
					)

					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Finished(
							"file atime is too new", pf)
					}

					outHolder.AppendToOut(pf)

					return
				}
			}

			if !o.Params.MinMtime.IsZero() {
				if timeStat.Mtime.Before(o.Params.MinMtime) {
					pf.SetFileProcessingError(
						NewFileMtimeIsTooOldError(o.Params.MinMtime, fileInfo.ModTime()),
					)

					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Finished(
							"file mtime is too old", pf)
					}

					outHolder.AppendToOut(pf)

					return
				}
			}

			if !o.Params.MaxMtime.IsZero() {
				if timeStat.Mtime.After(o.Params.MaxMtime) {
					pf.SetFileProcessingError(
						NewFileMtimeIsTooNewError(o.Params.MaxMtime, fileInfo.ModTime()),
					)

					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Finished(
							"file mtime is too new", pf)
					}

					outHolder.AppendToOut(pf)

					return
				}
			}

			if !o.Params.MinCtime.IsZero() {
				if timeStat.Ctime.Before(o.Params.MinCtime) {
					pf.SetFileProcessingError(
						NewFileCtimeIsTooOldError(o.Params.MinCtime, fileInfo.ModTime()),
					)

					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Finished(
							"file ctime is too old", pf)
					}

					outHolder.AppendToOut(pf)

					return
				}
			}

			if !o.Params.MaxCtime.IsZero() {
				if timeStat.Ctime.After(o.Params.MaxCtime) {
					pf.SetFileProcessingError(
						NewFileCtimeIsTooNewError(o.Params.MaxCtime, fileInfo.ModTime()),
					)

					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Finished(
							"file ctime is too new", pf)
					}

					outHolder.AppendToOut(pf)

					return
				}
			}

			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Finished("file time is valid", pf)
			}

			outHolder.AppendToOut(pf)
		}(pf)
	}

	wg.Wait()

	return outHolder.Out, nil
}

func (o *FileTimeValidateOperation) initTimeStatProvider() error {
	o.TimeStatProvider = &filetime.PlatformTimeStatProvider{}

	return nil
}

func (o *FileTimeValidateOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *FileTimeValidateOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}

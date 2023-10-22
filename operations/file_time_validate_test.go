package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"os"
	"testing"
	"time"
)

type testTimeStatProvider struct {
	atime time.Time
	mtime time.Time
	ctime time.Time
}

func (t *testTimeStatProvider) TimeStat(fi os.FileInfo) (*TimeStat, error) {
	return &TimeStat{
		Atime: t.atime,
		Mtime: t.mtime,
		Ctime: t.ctime,
	}, nil
}

func TestFileTimeValidateOperation_Handle(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file1KbBin, file1KbBinErr := capyfs.Filesystem.Open("testdata/file_1kb.bin")
	if file1KbBinErr != nil {
		t.Fatal(file1KbBinErr)
	}

	tests := []struct {
		name      string
		operation *FileTimeValidateOperation
		in        []files.ProcessableFile
		outAssert func(t *testing.T, out []files.ProcessableFile)
	}{
		{
			name: "should return file processing error when file atime is too old",
			operation: &FileTimeValidateOperation{
				Name: "file_time_validate",
				Params: &FileTimeValidateOperationParams{
					MinAtime: time.Now().Add(-time.Hour * 24 * 3),
					MaxAtime: time.Now().Add(-time.Hour * 24 * 1),
				},
				TimeStatProvider: &testTimeStatProvider{
					atime: time.Now().Add(-time.Hour * 24 * 4),
				},
			},
			in: []files.ProcessableFile{
				files.NewProcessableFile(file1KbBin),
			},
			outAssert: func(t *testing.T, out []files.ProcessableFile) {
				if len(out) != 1 {
					t.Fatalf("expected 1 files to be returned, but got %d", len(out))
				}

				if out[0].FileProcessingError == nil {
					t.Fatalf("expected file processing error, but got nil")
				}

				if out[0].FileProcessingError.Code() != ErrorCodeFileAtimeIsTooOld {
					t.Fatalf(
						"expected file processing error code %s, but got %s",
						ErrorCodeFileAtimeIsTooOld,
						out[0].FileProcessingError.Code(),
					)
				}
			},
		},
		{
			name: "should return file processing error when file atime is too new",
			operation: &FileTimeValidateOperation{
				Name: "file_time_validate",
				Params: &FileTimeValidateOperationParams{
					MinAtime: time.Now().Add(-time.Hour * 24 * 3),
					MaxAtime: time.Now().Add(-time.Hour * 24 * 1),
				},
				TimeStatProvider: &testTimeStatProvider{
					atime: time.Now().Add(-time.Hour * 12),
				},
			},
			in: []files.ProcessableFile{
				files.NewProcessableFile(file1KbBin),
			},
			outAssert: func(t *testing.T, out []files.ProcessableFile) {
				if len(out) != 1 {
					t.Fatalf("expected 1 files to be returned, but got %d", len(out))
				}

				if out[0].FileProcessingError == nil {
					t.Fatalf("expected file processing error, but got nil")
				}

				if out[0].FileProcessingError.Code() != ErrorCodeFileAtimeIsTooNew {
					t.Fatalf(
						"expected file processing error code %s, but got %s",
						ErrorCodeFileAtimeIsTooNew,
						out[0].FileProcessingError.Code(),
					)
				}
			},
		},
		{
			name: "should return file processing error when file mtime is too old",
			operation: &FileTimeValidateOperation{
				Name: "file_time_validate",
				Params: &FileTimeValidateOperationParams{
					MinMtime: time.Now().Add(-time.Hour * 24 * 3),
					MaxMtime: time.Now().Add(-time.Hour * 24 * 1),
				},
				TimeStatProvider: &testTimeStatProvider{
					mtime: time.Now().Add(-time.Hour * 24 * 4),
				},
			},
			in: []files.ProcessableFile{
				files.NewProcessableFile(file1KbBin),
			},
			outAssert: func(t *testing.T, out []files.ProcessableFile) {
				if len(out) != 1 {
					t.Fatalf("expected 1 files to be returned, but got %d", len(out))
				}

				if out[0].FileProcessingError == nil {
					t.Fatalf("expected file processing error, but got nil")
				}

				if out[0].FileProcessingError.Code() != ErrorCodeFileMtimeIsTooOld {
					t.Fatalf(
						"expected file processing error code %s, but got %s",
						ErrorCodeFileMtimeIsTooOld,
						out[0].FileProcessingError.Code(),
					)
				}
			},
		},
		{
			name: "should return file processing error when file mtime is too new",
			operation: &FileTimeValidateOperation{
				Name: "file_time_validate",
				Params: &FileTimeValidateOperationParams{
					MinMtime: time.Now().Add(-time.Hour * 24 * 3),
					MaxMtime: time.Now().Add(-time.Hour * 24 * 1),
				},
				TimeStatProvider: &testTimeStatProvider{
					mtime: time.Now().Add(-time.Hour * 12),
				},
			},
			in: []files.ProcessableFile{
				files.NewProcessableFile(file1KbBin),
			},
			outAssert: func(t *testing.T, out []files.ProcessableFile) {
				if len(out) != 1 {
					t.Fatalf("expected 1 files to be returned, but got %d", len(out))
				}

				if out[0].FileProcessingError == nil {
					t.Fatalf("expected file processing error, but got nil")
				}

				if out[0].FileProcessingError.Code() != ErrorCodeFileMtimeIsTooNew {
					t.Fatalf(
						"expected file processing error code %s, but got %s",
						ErrorCodeFileMtimeIsTooNew,
						out[0].FileProcessingError.Code(),
					)
				}
			},
		},
		{
			name: "should return file processing error when file ctime is too old",
			operation: &FileTimeValidateOperation{
				Name: "file_time_validate",
				Params: &FileTimeValidateOperationParams{
					MinCtime: time.Now().Add(-time.Hour * 24 * 3),
					MaxCtime: time.Now().Add(-time.Hour * 24 * 1),
				},
				TimeStatProvider: &testTimeStatProvider{
					ctime: time.Now().Add(-time.Hour * 24 * 4),
				},
			},
			in: []files.ProcessableFile{
				files.NewProcessableFile(file1KbBin),
			},
			outAssert: func(t *testing.T, out []files.ProcessableFile) {
				if len(out) != 1 {
					t.Fatalf("expected 1 files to be returned, but got %d", len(out))
				}

				if out[0].FileProcessingError == nil {
					t.Fatalf("expected file processing error, but got nil")
				}

				if out[0].FileProcessingError.Code() != ErrorCodeFileCtimeIsTooOld {
					t.Fatalf(
						"expected file processing error code %s, but got %s",
						ErrorCodeFileCtimeIsTooOld,
						out[0].FileProcessingError.Code(),
					)
				}
			},
		},
		{
			name: "should return file processing error when file ctime is too new",
			operation: &FileTimeValidateOperation{
				Name: "file_time_validate",
				Params: &FileTimeValidateOperationParams{
					MinCtime: time.Now().Add(-time.Hour * 24 * 3),
					MaxCtime: time.Now().Add(-time.Hour * 24 * 1),
				},
				TimeStatProvider: &testTimeStatProvider{
					ctime: time.Now().Add(-time.Hour * 12),
				},
			},
			in: []files.ProcessableFile{
				files.NewProcessableFile(file1KbBin),
			},
			outAssert: func(t *testing.T, out []files.ProcessableFile) {
				if len(out) != 1 {
					t.Fatalf("expected 1 files to be returned, but got %d", len(out))
				}

				if out[0].FileProcessingError == nil {
					t.Fatalf("expected file processing error, but got nil")
				}

				if out[0].FileProcessingError.Code() != ErrorCodeFileCtimeIsTooNew {
					t.Fatalf(
						"expected file processing error code %s, but got %s",
						ErrorCodeFileCtimeIsTooNew,
						out[0].FileProcessingError.Code(),
					)
				}
			},
		},
		{
			name: "should not return file processing error when file time stat is valid",
			operation: &FileTimeValidateOperation{
				Name: "file_time_validate",
				Params: &FileTimeValidateOperationParams{
					MinAtime: time.Now().Add(-time.Hour * 24 * 3),
					MaxAtime: time.Now().Add(-time.Hour * 24 * 1),
					MinMtime: time.Now().Add(-time.Hour * 24 * 3),
					MaxMtime: time.Now().Add(-time.Hour * 24 * 1),
					MinCtime: time.Now().Add(-time.Hour * 24 * 3),
					MaxCtime: time.Now().Add(-time.Hour * 24 * 1),
				},
				TimeStatProvider: &testTimeStatProvider{
					atime: time.Now().Add(-time.Hour * 24 * 2),
					mtime: time.Now().Add(-time.Hour * 24 * 2),
					ctime: time.Now().Add(-time.Hour * 24 * 2),
				},
			},
			in: []files.ProcessableFile{
				files.NewProcessableFile(file1KbBin),
			},
			outAssert: func(t *testing.T, out []files.ProcessableFile) {
				if len(out) != 1 {
					t.Fatalf("expected 1 files to be returned, but got %d", len(out))
				}

				if out[0].FileProcessingError != nil {
					t.Fatalf(
						"expected file processing error to be nil, but got %s",
						out[0].FileProcessingError.Code(),
					)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, err := test.operation.Handle(test.in, nil, nil)

			if err != nil {
				t.Fatalf("expected error to be nil, but got %s", err.Error())
			}

			test.outAssert(t, out)
		})
	}
}

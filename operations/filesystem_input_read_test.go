package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"testing"
)

func TestFilesystemInputReadOperation_HandleSingleFileRead(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	var in []files.ProcessableFile

	operation := &FilesystemInputReadOperation{
		Params: &FilesystemInputReadOperationParams{
			Target: "testdata/file_1kb.bin",
		},
	}
	out, err := operation.Handle(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 1 {
		t.Fatalf("len(out) = %d, want 1", len(out))
	}

	if out[0].FileProcessingError != nil {
		t.Fatalf(
			"FileProcessingError.Code() = %s, want nil",
			out[0].FileProcessingError.Code(),
		)
	}
}

func TestFilesystemInputReadOperation_HandleFileReadWithGlobPattern(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	var in []files.ProcessableFile

	operation := &FilesystemInputReadOperation{
		Params: &FilesystemInputReadOperationParams{
			Target: "testdata/*.bin",
		},
	}
	out, err := operation.Handle(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 3 {
		t.Fatalf("len(out) = %d, want 3", len(out))
	}

	for _, processableFile := range out {
		if processableFile.FileProcessingError != nil {
			t.Fatalf(
				"FileProcessingError.Code() = %s, want nil",
				processableFile.FileProcessingError.Code(),
			)
		}
	}
}

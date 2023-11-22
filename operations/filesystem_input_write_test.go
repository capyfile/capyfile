package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"testing"
)

func TestFilesystemInputWriteOperation_HandleSingleFileWriteWithOriginalFilename(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/image_512x512.png")
	if err != nil {
		t.Error(err)
	}

	in := []files.ProcessableFile{
		files.NewProcessableFile(file.Name()),
	}

	operation := &FilesystemInputWriteOperation{
		Params: &FilesystemInputWriteOperationParams{
			Destination:         "/tmp/testdata",
			UseOriginalFilename: true,
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

	exists, err := capyfs.FilesystemUtils.Exists("/tmp/testdata/image_512x512.png")
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Fatalf("file input has not been written to the destination")
	}
}

func TestFilesystemInputWriteOperation_HandleSingleFileWriteWithGeneratedFilename(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/image_512x512.png")
	if err != nil {
		t.Error(err)
	}

	in := []files.ProcessableFile{
		files.NewProcessableFile(file.Name()),
	}

	operation := &FilesystemInputWriteOperation{
		Params: &FilesystemInputWriteOperationParams{
			Destination:         "/tmp/testdata",
			UseOriginalFilename: false,
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

	exists, err := capyfs.FilesystemUtils.Exists("/tmp/testdata/" + out[0].GeneratedFilename())
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Fatalf("file input has not been written to the destination")
	}
}

func TestFilesystemInputWriteOperation_HandleMultipleFileWrite(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file1, err := capyfs.Filesystem.Open("testdata/image_512x512.png")
	if err != nil {
		t.Error(err)
	}
	file2, err := capyfs.Filesystem.Open("testdata/image_512x512.jpg")
	if err != nil {
		t.Error(err)
	}
	in := []files.ProcessableFile{
		files.NewProcessableFile(file1.Name()),
		files.NewProcessableFile(file2.Name()),
	}

	operation := &FilesystemInputWriteOperation{
		Params: &FilesystemInputWriteOperationParams{
			Destination:         "/tmp/testdata",
			UseOriginalFilename: true,
		},
	}
	out, err := operation.Handle(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want %d", len(out), len(in))
	}

	for _, processableFile := range out {
		if processableFile.FileProcessingError != nil {
			t.Fatalf(
				"FileProcessingError.Code() = %s, want nil",
				processableFile.FileProcessingError.Code(),
			)
		}

		exists, err := capyfs.FilesystemUtils.Exists("/tmp/testdata/" + processableFile.Filename())
		if err != nil {
			t.Error(err)
		}
		if !exists {
			t.Fatalf(
				"file input for %s has not been written to the destination",
				processableFile.OriginalFilename(),
			)
		}
	}
}

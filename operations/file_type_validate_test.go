package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"golang.org/x/exp/slices"
	"os"
	"testing"
)

func TestFileTypeValidateOperation_HandleFileOfAllowedType(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/image_512x512.jpg")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		files.NewProcessableFile(file.Name()),
	}

	operation := &FileTypeValidateOperation{
		Params: &FileTypeValidateOperationParams{
			AllowedMimeTypes: []string{"image/jpeg", "image/png", "image/webp"},
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

func TestFileTypeValidateOperation_HandleFilesOfAllowedType(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	osFileImageJpg, err := os.Open("testdata/image_512x512.jpg")
	if err != nil {
		t.Fatal(err)
	}
	osFileImagePng, err := os.Open("testdata/image_512x512.png")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		files.NewProcessableFile(osFileImageJpg.Name()),
		files.NewProcessableFile(osFileImagePng.Name()),
	}

	operation := &FileTypeValidateOperation{
		Params: &FileTypeValidateOperationParams{
			AllowedMimeTypes: []string{"image/jpeg", "image/png", "image/webp"},
		},
	}
	out, err := operation.Handle(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}

	for _, processableFile := range out {
		if processableFile.FileProcessingError != nil {
			t.Fatalf(
				"FileProcessingError.Code() = %s, want nil",
				out[0].FileProcessingError.Code(),
			)
		}
	}
}

func TestFileTypeValidateOperation_HandleFileOfNotAllowedType(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	osFileImagePng, err := os.Open("testdata/image_512x512.png")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		files.NewProcessableFile(osFileImagePng.Name()),
	}

	operation := &FileTypeValidateOperation{
		Params: &FileTypeValidateOperationParams{
			AllowedMimeTypes: []string{"image/jpeg", "image/webp"},
		},
	}
	out, err := operation.Handle(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 1 {
		t.Fatalf("len(out) = %d, want 1", len(out))
	}

	processableFile := out[0]

	if processableFile.FileProcessingError == nil {
		t.Fatalf("FileProcessingError = nil, want !nil")
	}

	if processableFile.FileProcessingError.Code() != ErrorCodeFileMimeTypeIsNotAllowed {
		t.Fatalf(
			"FileProcessingError.Code() = %s, want %s",
			out[0].FileProcessingError.Code(),
			ErrorCodeFileMimeTypeIsNotAllowed,
		)
	}
}

func TestFileTypeValidateOperation_HandleFilesOfNotAllowedType(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	osFileImageJpg, err := os.Open("testdata/image_512x512.jpg")
	if err != nil {
		t.Fatal(err)
	}
	osFileImagePng, err := os.Open("testdata/image_512x512.png")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		files.NewProcessableFile(osFileImageJpg.Name()),
		files.NewProcessableFile(osFileImagePng.Name()),
	}

	operation := &FileTypeValidateOperation{
		Params: &FileTypeValidateOperationParams{
			AllowedMimeTypes: []string{"image/webp"},
		},
	}
	out, err := operation.Handle(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}

	for _, processableFile := range out {
		if processableFile.FileProcessingError == nil {
			t.Fatalf("FileProcessingError = nil, want !nil")
		}

		if processableFile.FileProcessingError.Code() != ErrorCodeFileMimeTypeIsNotAllowed {
			t.Fatalf(
				"FileProcessingError.Code() = %s, want %s",
				processableFile.FileProcessingError.Code(),
				ErrorCodeFileMimeTypeIsNotAllowed,
			)
		}
	}
}

func TestFileTypeValidateOperation_HandleFilesOfAllowedAndNotAllowedTypes(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	osFileImageJpg, err := os.Open("testdata/image_512x512.jpg")
	if err != nil {
		t.Fatal(err)
	}
	osFileImageWebp, err := os.Open("testdata/image_512x512.webp")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		files.NewProcessableFile(osFileImageJpg.Name()),
		files.NewProcessableFile(osFileImageWebp.Name()),
	}

	operation := &FileTypeValidateOperation{
		Params: &FileTypeValidateOperationParams{
			AllowedMimeTypes: []string{"image/jpeg"},
		},
	}
	out, err := operation.Handle(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}

	processableFileImageJpgIdx := slices.IndexFunc(out, func(pf files.ProcessableFile) bool {
		return pf.Name() == osFileImageJpg.Name()
	})
	processableFileImageJpg := out[processableFileImageJpgIdx]
	if processableFileImageJpg.FileProcessingError != nil {
		t.Fatalf("FileProcessingError.Code() = %s, want nil", processableFileImageJpg.FileProcessingError.Code())
	}

	processableFileImageWebpIdx := slices.IndexFunc(out, func(pf files.ProcessableFile) bool {
		return pf.Name() == osFileImageWebp.Name()
	})
	processableFileImageWebp := out[processableFileImageWebpIdx]
	if processableFileImageWebp.FileProcessingError == nil {
		t.Fatalf("FileProcessingError = nil, want !nil")
	}
	if processableFileImageWebp.FileProcessingError.Code() != ErrorCodeFileMimeTypeIsNotAllowed {
		t.Fatalf(
			"FileProcessingError.Code() = %s, want %s",
			processableFileImageWebp.FileProcessingError.Code(),
			ErrorCodeFileMimeTypeIsNotAllowed,
		)
	}
}

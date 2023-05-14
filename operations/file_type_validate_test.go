package operations

import (
	"capyfile/files"
	"os"
	"testing"
)

func TestFileTypeValidateOperation_HandleFileOfAllowedType(t *testing.T) {
	osFile, err := os.Open("testdata/image_512x512.jpg")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: osFile},
	}

	operation := &FileTypeValidateOperation{
		Params: &FileTypeValidateOperationParams{
			AllowedMimeTypes: []string{"image/jpeg", "image/png", "image/webp"},
		},
	}
	out, err := operation.Handle(in)
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
	osFileImageJpg, err := os.Open("testdata/image_512x512.jpg")
	if err != nil {
		t.Fatal(err)
	}
	osFileImagePng, err := os.Open("testdata/image_512x512.png")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: osFileImageJpg},
		{File: osFileImagePng},
	}

	operation := &FileTypeValidateOperation{
		Params: &FileTypeValidateOperationParams{
			AllowedMimeTypes: []string{"image/jpeg", "image/png", "image/webp"},
		},
	}
	out, err := operation.Handle(in)
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
	osFileImagePng, err := os.Open("testdata/image_512x512.png")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: osFileImagePng},
	}

	operation := &FileTypeValidateOperation{
		Params: &FileTypeValidateOperationParams{
			AllowedMimeTypes: []string{"image/jpeg", "image/webp"},
		},
	}
	out, err := operation.Handle(in)
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
	osFileImageJpg, err := os.Open("testdata/image_512x512.jpg")
	if err != nil {
		t.Fatal(err)
	}
	osFileImagePng, err := os.Open("testdata/image_512x512.png")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: osFileImageJpg},
		{File: osFileImagePng},
	}

	operation := &FileTypeValidateOperation{
		Params: &FileTypeValidateOperationParams{
			AllowedMimeTypes: []string{"image/webp"},
		},
	}
	out, err := operation.Handle(in)
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
	osFileImageJpg, err := os.Open("testdata/image_512x512.jpg")
	if err != nil {
		t.Fatal(err)
	}
	osFileImageWebp, err := os.Open("testdata/image_512x512.webp")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: osFileImageJpg},
		{File: osFileImageWebp},
	}

	operation := &FileTypeValidateOperation{
		Params: &FileTypeValidateOperationParams{
			AllowedMimeTypes: []string{"image/jpeg"},
		},
	}
	out, err := operation.Handle(in)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}

	processableFileImageJpg := out[0]
	if processableFileImageJpg.FileProcessingError != nil {
		t.Fatalf("FileProcessingError.Code() = %s, want nil", processableFileImageJpg.FileProcessingError.Code())
	}

	processableFileImageWebp := out[1]
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

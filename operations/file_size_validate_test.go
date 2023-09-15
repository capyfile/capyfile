package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"os"
	"testing"
)

func TestFileSizeValidateOperation_HandleFileOfAllowedMaxSize(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/file_1kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: file},
	}

	operation := &FileSizeValidateOperation{
		Params: &FileSizeValidateOperationParams{
			MaxFileSize: 2048,
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

func TestFileSizeValidateOperation_HandleFileOfAllowedMinSize(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/file_1kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: file},
	}

	operation := &FileSizeValidateOperation{
		Params: &FileSizeValidateOperationParams{
			MinFileSize: 512,
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

func TestFileSizeValidateOperation_HandleFilesOfAllowedSize(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file1Kb, err := capyfs.Filesystem.Open("testdata/file_1kb.bin")
	if err != nil {
		t.Fatal(err)
	}
	file2Kb, err := capyfs.Filesystem.Open("testdata/file_2kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: file1Kb},
		{File: file2Kb},
	}

	operation := &FileSizeValidateOperation{
		Params: &FileSizeValidateOperationParams{
			MinFileSize: 512,
			MaxFileSize: 2048,
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
				processableFile.FileProcessingError.Code(),
			)
		}
	}
}

func TestFileSizeValidateOperation_HandleFileOfNotAllowedSize(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/file_2kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: file},
	}

	operation := &FileSizeValidateOperation{
		Params: &FileSizeValidateOperationParams{
			MaxFileSize: 1048,
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

	if processableFile.FileProcessingError.Code() != ErrorCodeFileSizeIsTooBig {
		t.Fatalf(
			"FileProcessingError.Code() = %s, want %s",
			processableFile.FileProcessingError.Code(),
			ErrorCodeFileSizeIsTooBig,
		)
	}
}

func TestFileSizeValidateOperation_HandleFileOfNotAllowedMinSize(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/file_1kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: file},
	}

	operation := &FileSizeValidateOperation{
		Params: &FileSizeValidateOperationParams{
			MinFileSize: 2048,
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

	if processableFile.FileProcessingError.Code() != ErrorCodeFileSizeIsTooSmall {
		t.Fatalf(
			"FileProcessingError.Code() = %s, want %s",
			processableFile.FileProcessingError.Code(),
			ErrorCodeFileSizeIsTooSmall,
		)
	}
}

func TestFileSizeValidateOperation_HandleFilesOfNotAllowedSize(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file1Kb, err := capyfs.Filesystem.Open("testdata/file_1kb.bin")
	if err != nil {
		t.Fatal(err)
	}
	file2Kb, err := capyfs.Filesystem.Open("testdata/file_2kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: file1Kb},
		{File: file2Kb},
	}

	operation := &FileSizeValidateOperation{
		Params: &FileSizeValidateOperationParams{
			MinFileSize: 1024 + 1,
			MaxFileSize: 2048 - 1,
		},
	}
	out, err := operation.Handle(in)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}

	processableFile1Kb := out[0]
	if processableFile1Kb.FileProcessingError == nil {
		t.Fatalf("FileProcessingError = nil, want !nil")
	}
	if processableFile1Kb.FileProcessingError.Code() != ErrorCodeFileSizeIsTooSmall {
		t.Fatalf(
			"FileProcessingError.Code() = %s, want %s",
			processableFile1Kb.FileProcessingError.Code(),
			ErrorCodeFileSizeIsTooSmall,
		)
	}

	processableFile2Kb := out[1]
	if processableFile2Kb.FileProcessingError == nil {
		t.Fatalf("FileProcessingError = nil, want !nil")
	}
	if processableFile2Kb.FileProcessingError.Code() != ErrorCodeFileSizeIsTooBig {
		t.Fatalf(
			"FileProcessingError.Code() = %s, want %s",
			processableFile2Kb.FileProcessingError.Code(),
			ErrorCodeFileSizeIsTooBig,
		)
	}
}

func TestFileSizeValidateOperation_HandleFilesOfAllowedAndNotAllowedSizes(t *testing.T) {
	osFile1Kb, err := os.Open("testdata/file_1kb.bin")
	if err != nil {
		t.Fatal(err)
	}
	osFile5Kb, err := os.Open("testdata/file_5kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		{File: osFile1Kb},
		{File: osFile5Kb},
	}

	operation := &FileSizeValidateOperation{
		Params: &FileSizeValidateOperationParams{
			MaxFileSize: 2048,
		},
	}
	out, err := operation.Handle(in)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}

	processableFile2Kb := out[0]
	if processableFile2Kb.FileProcessingError != nil {
		t.Fatalf("FileProcessingError.Code() = %s, want nil", processableFile2Kb.FileProcessingError.Code())
	}

	processableFile5Kb := out[1]
	if processableFile5Kb.FileProcessingError == nil {
		t.Fatalf("FileProcessingError = nil, want !nil")
	}
	if processableFile5Kb.FileProcessingError.Code() != ErrorCodeFileSizeIsTooBig {
		t.Fatalf(
			"FileProcessingError.Code() = %s, want %s",
			processableFile5Kb.FileProcessingError.Code(),
			ErrorCodeFileSizeIsTooBig,
		)
	}
}

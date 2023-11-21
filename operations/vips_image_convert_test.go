package operations

import (
	"capyfile/capyerr"
	"capyfile/capyfs"
	"capyfile/files"
	"errors"
	"testing"
)

func TestImageConvertOperation_HandlePngToJpgConversion(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/image_512x512.png")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		files.NewProcessableFile(file.Name()),
	}

	operation := &ImageConvertOperation{
		Params: &ImageConvertOperationParams{
			ToMimeType: "image/jpeg",
			Quality:    "best",
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

	if processableFile.FileProcessingError != nil {
		t.Fatalf(
			"FileProcessingError.Code() = %s, want nil",
			processableFile.FileProcessingError.Code(),
		)
	}

	mime, err := processableFile.Mime()
	if err != nil {
		t.Error(err)
	}

	if !mime.Is("image/jpeg") {
		t.Errorf("mime.Is(\"image/jpeg\") = false, want true")
	}
}

func TestImageConvertOperation_HandlePngToNotAllowedTypeConversion(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	file, err := capyfs.Filesystem.Open("testdata/image_512x512.png")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		files.NewProcessableFile(file.Name()),
	}

	operation := &ImageConvertOperation{
		Params: &ImageConvertOperationParams{
			ToMimeType: "image/vnd.adobe.photoshop",
			Quality:    "best",
		},
	}
	_, err = operation.Handle(in, nil, nil)
	if err == nil {
		t.Fatal("err = nil, want error")
	}

	var ocType *capyerr.OperationConfigurationType
	if errors.As(err, &ocType) {
		if ocType.Code() != ErrorCodeImageConvertOperationConfiguration {
			t.Fatalf(
				"expected error %s code, got error %s code",
				ErrorCodeImageConvertOperationConfiguration,
				ocType.Code(),
			)
		}
	} else {
		t.Fatalf("expected %v error, got %v error", ocType, err)
	}
}

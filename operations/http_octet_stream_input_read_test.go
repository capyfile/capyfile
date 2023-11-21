package operations

import (
	"bytes"
	"capyfile/capyfs"
	"capyfile/files"
	"net/http/httptest"
	"testing"
)

func TestHttpOctetStreamInputReadOperation_Handle(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	requestBody, rbReadErr := capyfs.FilesystemUtils.ReadFile("testdata/file_5kb.bin")
	if rbReadErr != nil {
		t.Fatal(rbReadErr)
	}

	r := httptest.NewRequest("POST", "http://example.com/foo", bytes.NewReader(requestBody))
	r.Header.Set("Content-Type", "application/octet-stream")

	operation := &HttpOctetStreamInputReadOperation{
		Req: r,
	}
	out, opHandleErr := operation.Handle([]files.ProcessableFile{}, nil, nil)
	if opHandleErr != nil {
		t.Fatal(opHandleErr)
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

	// Here we can check the file size to have some confidence that this is the same file
	// that we've read from the filesystem.

	originalFileStat, originalFileStatErr := capyfs.Filesystem.Stat("testdata/file_5kb.bin")
	if originalFileStatErr != nil {
		t.Fatal(originalFileStatErr)
	}

	uploadedFileStat, uploadedFileStatErr := capyfs.Filesystem.Stat(out[0].Name())
	if uploadedFileStatErr != nil {
		t.Fatal(uploadedFileStatErr)
	}

	if originalFileStat.Size() != uploadedFileStat.Size() {
		t.Fatalf(
			"uploadedFileStat.Size() = %d, want %d",
			uploadedFileStat.Size(),
			originalFileStat.Size(),
		)
	}
}

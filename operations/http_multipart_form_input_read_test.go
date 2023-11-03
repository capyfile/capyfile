package operations

import (
	"bytes"
	"capyfile/capyfs"
	"capyfile/files"
	"mime/multipart"
	"net/http/httptest"
	"testing"
)

func TestHttpMultipartFormInputReadOperation_Handle(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	requestBody := new(bytes.Buffer)
	w := multipart.NewWriter(requestBody)

	formFile1Kb, formFile1KbErr := w.CreateFormFile("file1kb", "file_1kb.bin")
	if formFile1KbErr != nil {
		t.Fatal(formFile1KbErr)
	}
	file1KbBytes, file1KbReadErr := capyfs.FilesystemUtils.ReadFile("testdata/file_1kb.bin")
	if file1KbReadErr != nil {
		t.Fatal(file1KbReadErr)
	}
	_, file1KbWriteErr := formFile1Kb.Write(file1KbBytes)
	if file1KbWriteErr != nil {
		t.Fatal(file1KbWriteErr)
	}

	formFile2Kb, formFile2KbErr := w.CreateFormFile("file2kb", "file_2kb.bin")
	if formFile2KbErr != nil {
		t.Fatal(formFile2KbErr)
	}
	file2KbBytes, file2KbReadErr := capyfs.FilesystemUtils.ReadFile("testdata/file_2kb.bin")
	if file2KbReadErr != nil {
		t.Fatal(file2KbReadErr)
	}
	_, file2KbWriteErr := formFile2Kb.Write(file2KbBytes)
	if file2KbWriteErr != nil {
		t.Fatal(file2KbWriteErr)
	}

	wCloseErr := w.Close()
	if wCloseErr != nil {
		t.Fatal(wCloseErr)
	}

	r := httptest.NewRequest("POST", "http://example.com/foo", requestBody)
	r.Header.Set("Content-Type", w.FormDataContentType())

	operation := &HttpMultipartFormInputReadOperation{
		Req: r,
	}
	out, opHandleErr := operation.Handle([]files.ProcessableFile{}, nil, nil)
	if opHandleErr != nil {
		t.Fatal(opHandleErr)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}

	for _, pf := range out {
		if pf.OriginalFilename() != "file_1kb.bin" && pf.OriginalFilename() != "file_2kb.bin" {
			t.Fatalf(
				"pf.OriginalFilename() = %s, want file_1kb.bin or file_2kb.bin",
				pf.OriginalFilename(),
			)
		}

		if pf.FileProcessingError != nil {
			t.Fatalf(
				"FileProcessingError.Code() = %s, want nil",
				pf.FileProcessingError.Code(),
			)
		}
	}
}

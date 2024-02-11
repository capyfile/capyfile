package main

import (
	"capyfile/capyfs"
	"capyfile/capysvc"
	"capyfile/capysvc/common"
	"capyfile/parameters"
	"encoding/json"
	"golang.org/x/exp/rand"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testServiceDefinition is a test service definition that just checks the file size.
var testServiceDefinition = capysvc.Service{
	Name: "validator",
	Processors: []capysvc.Processor{
		{
			Name: "bin_file",
			Operations: []capysvc.Operation{
				{
					Name: "http_multipart_form_input_read",
				},
				{
					Name: "file_size_validate",
					Params: map[string]parameters.Parameter{
						"maxFileSize": {
							SourceType: "value",
							Source:     float64(5 * 1024), // max file size is 5kb
						},
					},
				},
			},
		},
	},
}

func TestProcessFiles(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	testLoggerInitErr := common.InitTestLogger()
	if testLoggerInitErr != nil {
		t.Errorf("expected error to be nil, got %v", testLoggerInitErr)
	}

	capysvc.LoadTestServiceDefinition(&testServiceDefinition)

	pipeReader, pipeWriter := io.Pipe()
	multipartWriter := multipart.NewWriter(pipeWriter)

	// Write the multipart form headers that contains random bytes.
	// Prat1 is 5kb bin file (valid) and part2 is 7kb bin file (invalid).
	go func() {
		defer multipartWriter.Close()

		part1, ffErr := multipartWriter.CreateFormFile("file1", "file_5kb.bin")
		if ffErr != nil {
			t.Errorf("expected error to be nil, got %v", ffErr)
		}

		buf1 := make([]byte, 5*1024)

		_, readErr1 := rand.Read(buf1)
		if readErr1 != nil {
			t.Errorf("expected error to be nil, got %v", readErr1)
		}

		_, writeErr1 := part1.Write(buf1)
		if writeErr1 != nil {
			t.Errorf("expected error to be nil, got %v", writeErr1)
		}

		part2, ffErr2 := multipartWriter.CreateFormFile("file2", "file_7kb.bin")
		if ffErr2 != nil {
			t.Errorf("expected error to be nil, got %v", ffErr2)
		}

		buf2 := make([]byte, 7*1024)

		_, readErr2 := rand.Read(buf2)
		if readErr2 != nil {
			t.Errorf("expected error to be nil, got %v", readErr2)
		}

		_, writeErr2 := part2.Write(buf2)
		if writeErr2 != nil {
			t.Errorf("expected error to be nil, got %v", writeErr2)
		}
	}()

	req := httptest.NewRequest(http.MethodPost, "/validator/bin_file", pipeReader)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	w := httptest.NewRecorder()

	s := Server{
		Concurrency:     true,
		ConcurrencyMode: "event",
	}
	s.Handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	jsonData, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}

	var data map[string]interface{}
	jsonErr := json.Unmarshal(jsonData, &data)
	if jsonErr != nil {
		t.Errorf("expected error to be nil, got %v", jsonErr)
	}

	if data["status"] != "PARTIAL" {
		t.Errorf("expected status to be PARTIAL, got %v", data["status"])
	}

	files := data["files"].([]interface{})
	if len(files) != 1 {
		t.Errorf("expected files length to be 1, got %v", len(files))
	}

	errors := data["errors"].([]interface{})
	if len(errors) != 1 {
		t.Errorf("expected errors length to be 1, got %v", len(errors))
	}
}

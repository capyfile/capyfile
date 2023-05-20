package capysvc

import (
	"bytes"
	"capyfile/capyfs"
	"capyfile/capysvc/service"
	"capyfile/files"
	"capyfile/operations"
	"net/http"
	"testing"
)

var testServiceDefinition = Service{
	Name: "validator",
	Processors: []Processor{
		{
			Name: "photo",
			Operations: []Operation{
				{
					Name: "file_size_validate",
					Params: map[string]OperationParameter{
						"maxFileSize": {
							SourceType: "value",
							Source:     float64(1 >> 20),
						},
					},
				},
				{
					Name: "file_type_validate",
					Params: map[string]OperationParameter{
						"allowedMimeTypes": {
							SourceType: "http_post",
							Source:     "allowed_mime_types",
						},
					},
				},
			},
		},
	},
}

func TestService_Processor(t *testing.T) {
	var p *Processor

	p = testServiceDefinition.Processor("photo")
	if p == nil {
		t.Fatalf("Service.Processor(photo) = nil, want photo processor")
	}

	p = testServiceDefinition.Processor("video")
	if p != nil {
		t.Fatalf("Service.Processor(video) != nil, want nil")
	}
}

func TestService_RunProcessors(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	imageFile, err := capyfs.Filesystem.Open("testdata/image_512x512.jpg")
	if err != nil {
		t.Fatal(err)
	}
	binFile, err := capyfs.Filesystem.Open("testdata/file_5kb.bin")
	if err != nil {
		t.Fatal(err)
	}

	in := []files.ProcessableFile{
		*files.NewProcessableFile(imageFile),
		*files.NewProcessableFile(binFile),
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://example.com/validator/photo",
		bytes.NewReader([]byte("allowed_mime_types=[\"image/jpeg\"]")))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		t.Fatalf("expect no error while creating request, got %v", err)
	}

	out, err := testServiceDefinition.RunProcessor(
		&service.ServerContext{Req: req},
		"photo",
		in)

	if err != nil {
		t.Fatalf("expect no error while running processor, got %v", err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}

	imageProcessableFile := out[0]
	if imageProcessableFile.HasFileProcessingError() {
		t.Fatalf("expect no error for image processable file, got %v", imageProcessableFile.FileProcessingError)
	}

	binProcessableFile := out[1]
	if !binProcessableFile.HasFileProcessingError() {
		t.Fatalf("expect an error for bin processable file, got nil")
	}
	if binProcessableFile.FileProcessingError.Code() != operations.ErrorCodeFileMimeTypeIsNotAllowed {
		t.Fatalf(
			"expect %s error code for bin processable file, got %s",
			operations.ErrorCodeFileMimeTypeIsNotAllowed,
			binProcessableFile.FileProcessingError.Code())
	}
}

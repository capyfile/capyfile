package capysvc

import (
	"bytes"
	"capyfile/capyfs"
	"capyfile/files"
	"capyfile/operations"
	"golang.org/x/exp/slices"
	"net/http"
	"testing"
)

func testServiceDefinitionForServerContext() Service {
	return Service{
		Name: "validator",
		Processors: []Processor{
			{
				Name: "photo",
				Operations: []Operation{
					{
						Name: "file_size_validate",
						Params: map[string]OperationParameter{
							"minFileSize": {
								SourceType: "value",
								Source:     float64(1),
							},
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
}

func testServiceDefinitionForCliContext() Service {
	return Service{
		Name: "validator",
		Processors: []Processor{
			{
				Name: "photo",
				Operations: []Operation{
					{
						Name: "file_size_validate",
						Params: map[string]OperationParameter{
							"minFileSize": {
								SourceType: "value",
								Source:     float64(1),
							},
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
								SourceType: "value",
								Source:     []string{"image/jpeg"},
							},
						},
					},
				},
			},
		},
	}
}

func TestService_Processor(t *testing.T) {
	var p *Processor

	sd := testServiceDefinitionForServerContext()

	p = sd.FindProcessor("photo")
	if p == nil {
		t.Fatalf("Service.FindProcessor(photo) = nil, want photo processor")
	}

	p = sd.FindProcessor("video")
	if p != nil {
		t.Fatalf("Service.FindProcessor(video) != nil, want nil")
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
		files.NewProcessableFile(imageFile),
		files.NewProcessableFile(binFile),
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://example.com/validator/photo",
		bytes.NewReader([]byte("allowed_mime_types=[\"image/jpeg\"]")))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		t.Fatalf("expect no error while creating request, got %v", err)
	}

	sd := testServiceDefinitionForServerContext()

	out, err := sd.RunProcessor(
		NewServerContext(req, nil),
		"photo",
		in)

	if err != nil {
		t.Fatalf("expect no error while running processor, got %v", err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}

	imageProcessableFileIdx := slices.IndexFunc(out, func(pf files.ProcessableFile) bool {
		return pf.File.Name() == imageFile.Name()
	})
	imageProcessableFile := out[imageProcessableFileIdx]
	if imageProcessableFile.HasFileProcessingError() {
		t.Fatalf("expect no error for image processable file, got %v", imageProcessableFile.FileProcessingError)
	}

	binProcessableFileIdx := slices.IndexFunc(out, func(pf files.ProcessableFile) bool {
		return pf.File.Name() == binFile.Name()
	})
	binProcessableFile := out[binProcessableFileIdx]
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

func TestService_RunProcessorsConcurrentlyWithServerContext(t *testing.T) {
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
		files.NewProcessableFile(imageFile),
		files.NewProcessableFile(binFile),
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://example.com/validator/photo",
		bytes.NewReader([]byte("allowed_mime_types=[\"image/jpeg\"]")))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		t.Fatalf("expect no error while creating request, got %v", err)
	}

	sd := testServiceDefinitionForServerContext()

	out, err := sd.RunProcessorConcurrently(
		NewServerContext(req, nil),
		"photo",
		in,
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("expect no error while running processor, got %v", err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}

	imageProcessableFileIdx := slices.IndexFunc(out, func(pf files.ProcessableFile) bool {
		return pf.File.Name() == imageFile.Name()
	})
	imageProcessableFile := out[imageProcessableFileIdx]
	if imageProcessableFile.HasFileProcessingError() {
		t.Fatalf("expect no error for image processable file, got %v", imageProcessableFile.FileProcessingError)
	}

	binProcessableFileIdx := slices.IndexFunc(out, func(pf files.ProcessableFile) bool {
		return pf.File.Name() == binFile.Name()
	})
	binProcessableFile := out[binProcessableFileIdx]
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

func TestService_RunProcessorsConcurrentlyWithCliContext(t *testing.T) {
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
		files.NewProcessableFile(imageFile),
		files.NewProcessableFile(binFile),
	}

	errorCh := make(chan operations.OperationError)
	notificationCh := make(chan operations.OperationNotification)

	go func() {
		for {
			select {
			case <-errorCh:
			case <-notificationCh:
			}
		}
	}()

	sd := testServiceDefinitionForCliContext()

	out, err := sd.RunProcessorConcurrently(
		NewCliContext(),
		"photo",
		in,
		errorCh,
		notificationCh,
	)

	if err != nil {
		t.Fatalf("expect no error while running processor, got %v", err)
	}

	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}

	imageProcessableFileIdx := slices.IndexFunc(out, func(pf files.ProcessableFile) bool {
		return pf.File.Name() == imageFile.Name()
	})
	imageProcessableFile := out[imageProcessableFileIdx]
	if imageProcessableFile.HasFileProcessingError() {
		t.Fatalf("expect no error for image processable file, got %v", imageProcessableFile.FileProcessingError)
	}

	binProcessableFileIdx := slices.IndexFunc(out, func(pf files.ProcessableFile) bool {
		return pf.File.Name() == binFile.Name()
	})
	binProcessableFile := out[binProcessableFileIdx]
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

func TestService_RunProcessorsConcurrentlyWithEmptyInput(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	sd := testServiceDefinitionForCliContext()

	out, err := sd.RunProcessorConcurrently(
		NewCliContext(),
		"photo",
		[]files.ProcessableFile{},
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("expect no error while running processor, got %v", err)
	}

	if len(out) != 0 {
		t.Fatalf("len(out) = %d, want 0", len(out))
	}
}

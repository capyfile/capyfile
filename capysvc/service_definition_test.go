package capysvc

import (
	"bytes"
	"capyfile/capyfs"
	"capyfile/files"
	"capyfile/operations"
	"capyfile/parameters"
	"fmt"
	"golang.org/x/exp/rand"
	"golang.org/x/exp/slices"
	"net/http"
	"testing"
)

func benchmarkServiceDefinition() Service {
	return Service{
		Name: "bin_files",
		Processors: []Processor{
			{
				Name: "validate",
				Operations: []Operation{
					{
						Name: "filesystem_input_read",
						Params: map[string]parameters.Parameter{
							"target": {
								SourceType: "value",
								Source:     "/tmp/testdata/*",
							},
						},
					},
					{
						Name: "file_size_validate",
						Params: map[string]parameters.Parameter{
							"maxFileSize": {
								SourceType: "value",
								Source:     5 * 1024, // max file size is 5kb
							},
						},
					},
					{
						Name: "file_type_validate",
						Params: map[string]parameters.Parameter{
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

func BenchmarkService_RunProcessorConcurrently(b *testing.B) {
	capyfs.InitCopyOnWriteFilesystem()

	sizes := []int{1, 3, 5, 7, 10}
	filesPerSize := 100

	mkdirErr := capyfs.FilesystemUtils.MkdirAll("/tmp/testdata", 0755)
	if mkdirErr != nil {
		b.Fatalf("expected error to be nil, got %v", mkdirErr)
	}

	for _, sizeKb := range sizes {
		for i := 0; i < filesPerSize; i++ {
			buf := make([]byte, sizeKb*1024)
			_, readErr := rand.Read(buf)
			if readErr != nil {
				b.Fatalf("expected error to be nil, got %v", readErr)
			}

			fileWriteErr := capyfs.FilesystemUtils.WriteFile(
				fmt.Sprintf("/tmp/testdata/file_%dkb_%d.bin", sizeKb, i),
				buf,
				0644)
			if fileWriteErr != nil {
				b.Fatalf("expected error to be nil, got %v", fileWriteErr)
			}
		}
	}

	sd := benchmarkServiceDefinition()

	for i := 0; i < b.N; i++ {
		procOut, procErr := sd.RunProcessorConcurrently(
			NewCliContext(),
			"validate",
			[]files.ProcessableFile{},
			nil,
			nil,
		)
		if procErr != nil {
			b.Fatalf("expected error to be nil, got %v", procErr)
		}

		if len(procOut) != len(sizes)*filesPerSize {
			b.Fatalf("len(procOut) = %d, want %d", len(procOut), len(sizes)*filesPerSize)
		}
	}
}

func testServiceDefinitionForServerContext() Service {
	return Service{
		Name: "validator",
		Processors: []Processor{
			{
				Name: "photo",
				Operations: []Operation{
					{
						Name: "file_size_validate",
						Params: map[string]parameters.Parameter{
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
						Params: map[string]parameters.Parameter{
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
						Params: map[string]parameters.Parameter{
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
						Params: map[string]parameters.Parameter{
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
		files.NewProcessableFile(imageFile.Name()),
		files.NewProcessableFile(binFile.Name()),
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
		return pf.Name() == imageFile.Name()
	})
	imageProcessableFile := out[imageProcessableFileIdx]
	if imageProcessableFile.HasFileProcessingError() {
		t.Fatalf("expect no error for image processable file, got %v", imageProcessableFile.FileProcessingError)
	}

	binProcessableFileIdx := slices.IndexFunc(out, func(pf files.ProcessableFile) bool {
		return pf.Name() == binFile.Name()
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
		files.NewProcessableFile(imageFile.Name()),
		files.NewProcessableFile(binFile.Name()),
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
		return pf.Name() == imageFile.Name()
	})
	imageProcessableFile := out[imageProcessableFileIdx]
	if imageProcessableFile.HasFileProcessingError() {
		t.Fatalf("expect no error for image processable file, got %v", imageProcessableFile.FileProcessingError)
	}

	binProcessableFileIdx := slices.IndexFunc(out, func(pf files.ProcessableFile) bool {
		return pf.Name() == binFile.Name()
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
		files.NewProcessableFile(imageFile.Name()),
		files.NewProcessableFile(binFile.Name()),
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
		return pf.Name() == imageFile.Name()
	})
	imageProcessableFile := out[imageProcessableFileIdx]
	if imageProcessableFile.HasFileProcessingError() {
		t.Fatalf("expect no error for image processable file, got %v", imageProcessableFile.FileProcessingError)
	}

	binProcessableFileIdx := slices.IndexFunc(out, func(pf files.ProcessableFile) bool {
		return pf.Name() == binFile.Name()
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

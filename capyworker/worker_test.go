package main

import (
	"capyfile/capyfs"
	"capyfile/capysvc"
	"capyfile/capysvc/common"
	"capyfile/parameters"
	"fmt"
	"golang.org/x/exp/rand"
	"testing"
)

// testServiceDefinition is a test service definition that just checks the file size
// and removes the files that are greater than 5kb.
var testServiceDefinition = capysvc.Service{
	Name: "bin_files",
	Processors: []capysvc.Processor{
		{
			Name: "remove_large_files",
			Operations: []capysvc.Operation{
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
					Name:        "filesystem_input_remove",
					TargetFiles: "with_errors",
				},
			},
		},
	},
}

func TestWorker_Run(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	testLoggerInitErr := common.InitTestLogger()
	if testLoggerInitErr != nil {
		t.Fatalf("expected error to be nil, got %v", testLoggerInitErr)
	}

	capysvc.LoadTestServiceDefinition(&testServiceDefinition)

	// Create some test files of different sizes in /tmp/testdata directory.

	mkdirErr := capyfs.FilesystemUtils.MkdirAll("/tmp/testdata", 0755)
	if mkdirErr != nil {
		t.Fatalf("expected error to be nil, got %v", mkdirErr)
	}

	for _, sizeKb := range []int{1, 3, 5, 7, 10} {
		buf := make([]byte, sizeKb*1024)
		_, readErr := rand.Read(buf)
		if readErr != nil {
			t.Fatalf("expected error to be nil, got %v", readErr)
		}

		fileWriteErr := capyfs.FilesystemUtils.WriteFile(
			fmt.Sprintf("/tmp/testdata/file_%dkb", sizeKb)+".bin",
			buf,
			0644)
		if fileWriteErr != nil {
			t.Fatalf("expected error to be nil, got %v", fileWriteErr)
		}
	}

	worker := Worker{
		MaxIterations: 3,
	}
	runError := worker.Run("bin_files:remove_large_files")
	if runError != nil {
		t.Fatalf("expected error to be nil, got %v", runError)
	}

	// According to the testServiceDefinition, the files that are greater than 5kb
	// should be removed.

	for _, sizeKb := range []int{1, 3, 5} {
		fileExists, fileExistsErr := capyfs.FilesystemUtils.Exists(
			fmt.Sprintf("/tmp/testdata/file_%dkb", sizeKb) + ".bin")
		if fileExistsErr != nil {
			t.Fatalf("expected error to be nil, got %v", fileExistsErr)
		}
		if !fileExists {
			t.Fatalf("expected file of %dkb to exist, got false", sizeKb)
		}
	}

	for _, sizeKb := range []int{7, 10} {
		fileExists, fileExistsErr := capyfs.FilesystemUtils.Exists(
			fmt.Sprintf("/tmp/testdata/file_%dkb", sizeKb) + ".bin")
		if fileExistsErr != nil {
			t.Fatalf("expected error to be nil, got %v", fileExistsErr)
		}
		if fileExists {
			t.Fatalf("expected file of %dkb to not exist, got true", sizeKb)
		}
	}
}

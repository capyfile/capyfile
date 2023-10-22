package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"os"
	"testing"
)

func TestFilesystemInputRemoveOperation_HandleFileRemove(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	mkDirErr := capyfs.FilesystemUtils.MkdirAll("/tmp/testdata", os.ModePerm)
	if mkDirErr != nil {
		t.Fatal(mkDirErr)
	}

	file1, wErr1 := capyfs.FilesystemUtils.Create("/tmp/testdata/file1.txt")
	if wErr1 != nil {
		t.Fatal(wErr1)
	}
	file2, wErr2 := capyfs.FilesystemUtils.Create("/tmp/testdata/file2.txt")
	if wErr2 != nil {
		t.Fatal(wErr2)
	}

	in := []files.ProcessableFile{
		{
			File: file1,
			Metadata: &files.ProcessableFileMetadata{
				OriginalFilename: "file1.txt",
			},
		},
		{
			File: file2,
			Metadata: &files.ProcessableFileMetadata{
				OriginalFilename: "file2.txt",
			},
		},
	}

	operation := &FilesystemInputRemoveOperation{}
	out, err := operation.Handle(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 0 {
		t.Fatalf("len(out) = %d, want 0", len(out))
	}

	exists1, err := capyfs.FilesystemUtils.Exists("/tmp/testdata/file1.txt")
	if err != nil {
		t.Fatal(err)
	}
	if exists1 {
		t.Fatalf("file input has not been removed")
	}

	exists2, err := capyfs.FilesystemUtils.Exists("/tmp/testdata/file2.txt")
	if err != nil {
		t.Fatal(err)
	}
	if exists2 {
		t.Fatalf("file input has not been removed")
	}
}

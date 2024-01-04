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
		files.NewProcessableFile(file1.Name()),
		files.NewProcessableFile(file2.Name()),
	}

	operation := &FilesystemInputRemoveOperation{
		Params: &FilesystemInputRemoveOperationParams{
			RemoveOriginalFile: false,
		},
	}
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

func TestFilesystemInputRemoveOperation_HandleFileRemoveWithOriginalFile(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	mkDirErr := capyfs.FilesystemUtils.MkdirAll("/tmp/testdata", os.ModePerm)
	if mkDirErr != nil {
		t.Fatal(mkDirErr)
	}

	origFile, origFileErr := capyfs.FilesystemUtils.Create("/tmp/testdata/file1.txt")
	if origFileErr != nil {
		t.Fatal(origFileErr)
	}
	file, fileErr := capyfs.FilesystemUtils.Create("/tmp/testdata/file2.txt")
	if fileErr != nil {
		t.Fatal(fileErr)
	}

	pf := files.NewProcessableFile(origFile.Name())
	pf.ReplaceFile(file.Name())

	in := []files.ProcessableFile{
		pf,
	}

	operation := &FilesystemInputRemoveOperation{
		Params: &FilesystemInputRemoveOperationParams{
			RemoveOriginalFile: true,
		},
	}
	out, err := operation.Handle(in, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(out) != 0 {
		t.Fatalf("len(out) = %d, want 0", len(out))
	}

	origFileExists, origFileExistsErr := capyfs.FilesystemUtils.Exists("/tmp/testdata/file1.txt")
	if origFileExistsErr != nil {
		t.Fatal(origFileExistsErr)
	}
	if origFileExists {
		t.Fatalf("original file input has not been removed")
	}

	fileExists, fileExistsErr := capyfs.FilesystemUtils.Exists("/tmp/testdata/file2.txt")
	if fileExistsErr != nil {
		t.Fatal(fileExistsErr)
	}
	if fileExists {
		t.Fatalf("file input has not been removed")
	}
}

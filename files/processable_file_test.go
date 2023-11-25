package files

import (
	"capyfile/capyfs"
	"testing"
)

func TestNewProcessableFile_ReplaceFile(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	origFile, origFileErr := capyfs.FilesystemUtils.TempFile("", "")
	if origFileErr != nil {
		t.Fatal(origFileErr)
	}

	pf := NewProcessableFile(origFile.Name())
	pf.RemoveOnFreeResources()

	fileForReplacement, fileForReplacementErr := capyfs.FilesystemUtils.TempFile("", "")
	if fileForReplacementErr != nil {
		t.Fatal(fileForReplacementErr)
	}

	pf.ReplaceFile(fileForReplacement.Name())

	// We want to make sure that the original origFile is presented in the filesystem,
	// and it is not removed during the cleanup.
	exists, existsErr := capyfs.FilesystemUtils.Exists(origFile.Name())
	if existsErr != nil {
		t.Fatal(existsErr)
	}

	if !exists {
		t.Fatalf("expected %s to exist, but it doesn't", origFile.Name())
	}
}

func TestProcessableFile_FreeResources(t *testing.T) {
	capyfs.InitCopyOnWriteFilesystem()

	origFile, origFileErr := capyfs.FilesystemUtils.TempFile("", "")
	if origFileErr != nil {
		t.Fatal(origFileErr)
	}

	// By default, the original file is preserved and the cleanup policy is set to "none".
	// So the file should be presented after the FreeResources() call.
	pf := NewProcessableFile(origFile.Name())

	frErr := pf.FreeResources()
	if frErr != nil {
		t.Fatal(frErr)
	}

	// The file should be presented because of the cleanup policy,
	// and the original file preservation.
	exists, existsErr := capyfs.FilesystemUtils.Exists(origFile.Name())
	if existsErr != nil {
		t.Fatal(existsErr)
	}

	if !exists {
		t.Fatalf("expected %s to exist, but it doesn't", origFile.Name())
	}

	// We changed the cleanup policy to "remove", but the file should still be presented
	// because of the original file preservation.
	pf.RemoveOnFreeResources()

	frErr = pf.FreeResources()
	if frErr != nil {
		t.Fatal(frErr)
	}

	// The file should be presented because of the original file preservation.
	exists, existsErr = capyfs.FilesystemUtils.Exists(origFile.Name())
	if existsErr != nil {
		t.Fatal(existsErr)
	}

	if !exists {
		t.Fatalf("expected %s to exist, but it doesn't", origFile.Name())
	}

	// Current cleanup policy is "remove". We are going to change the original
	//file preservation flag to false, so the file should be removed.
	pf.PreserveOriginalProcessableFile = false

	frErr = pf.FreeResources()
	if frErr != nil {
		t.Fatal(frErr)
	}

	// The file should be removed because of the original file is no longer preserved.
	exists, existsErr = capyfs.FilesystemUtils.Exists(origFile.Name())
	if existsErr != nil {
		t.Fatal(existsErr)
	}

	if exists {
		t.Fatalf("expected %s to not exist, but it does", origFile.Name())
	}
}

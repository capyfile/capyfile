package capyutils

import (
	"capyfile/capyfs"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/spf13/afero"
	"io"
	"os"
)

func prepareAppDirectory() error {
	homedir, uhdErr := os.UserHomeDir()
	if uhdErr != nil {
		return uhdErr
	}

	tmpMkDirErr := capyfs.FilesystemUtils.MkdirAll(homedir+"/.capyfile/tmp", 0755)
	if tmpMkDirErr != nil {
		return tmpMkDirErr
	}

	return nil
}

func WriteBytesToAppTmpDirectory(b []byte) (afero.File, error) {
	prepErr := prepareAppDirectory()
	if prepErr != nil {
		return nil, prepErr
	}

	homedir, uhdErr := os.UserHomeDir()
	if uhdErr != nil {
		return nil, uhdErr
	}

	file, fileErr := capyfs.Filesystem.Create(homedir + "/.capyfile/tmp/" + gonanoid.Must())
	if fileErr != nil {
		return nil, fileErr
	}
	defer file.Close()

	_, writeErr := file.Write(b)
	if writeErr != nil {
		return nil, writeErr
	}

	return file, nil
}

func WriteReaderToAppTmpDirectory(r io.Reader) (afero.File, error) {
	prepErr := prepareAppDirectory()
	if prepErr != nil {
		return nil, prepErr
	}

	homedir, uhdErr := os.UserHomeDir()
	if uhdErr != nil {
		return nil, uhdErr
	}

	file, fileErr := capyfs.Filesystem.Create(homedir + "/.capyfile/tmp/" + gonanoid.Must())
	if fileErr != nil {
		return nil, fileErr
	}
	defer file.Close()

	_, copyErr := io.Copy(file, r)
	if copyErr != nil {
		return nil, copyErr
	}

	return file, nil
}

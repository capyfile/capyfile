package capyutils

import (
	"capyfile/capyfs"
	"github.com/spf13/afero"
	"io"
)

func WriteBytesToTempFileAndLeaveOpen(b []byte) (file afero.File, err error) {
	file, err = capyfs.OsFilesystemUtils.TempFile("", "capyfile_")
	if err != nil {
		return nil, err
	}

	_, err = file.Write(b)
	if err != nil {
		return nil, err
	}

	return file, err
}

func WriteReaderToTempFileAndLeaveOpen(r io.Reader) (file afero.File, err error) {
	file, err = capyfs.OsFilesystemUtils.TempFile("", "capyfile_")
	if err != nil {
		return file, err
	}

	_, err = io.Copy(file, r)
	if err != nil {
		return file, err
	}

	return file, nil
}

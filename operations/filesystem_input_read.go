package operations

import (
	"capyfile/capyfs"
	"capyfile/files"
	"github.com/spf13/afero"
)

// FilesystemInputReadOperation reads input from the filesystem for further processing.
type FilesystemInputReadOperation struct {
	Params *FilesystemInputReadOperationParams
}

type FilesystemInputReadOperationParams struct {
	// Target is the target file or directory to read from. Can be a glob pattern.
	Target string
}

func (o *FilesystemInputReadOperation) Handle(in []files.ProcessableFile) ([]files.ProcessableFile, error) {
	matches, matchesErr := afero.Glob(capyfs.Filesystem, o.Params.Target)
	if matchesErr != nil {
		return in, matchesErr
	}

	for _, match := range matches {
		f, fileOpenErr := capyfs.Filesystem.Open(match)
		if fileOpenErr != nil {
			in = append(in, *files.NewUnreadableProcessableFile(match, NewFileInputIsUnreadableError(fileOpenErr)))

			continue
		}

		in = append(in, *files.NewProcessableFile(f))
	}

	return in, nil
}

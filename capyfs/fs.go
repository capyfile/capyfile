package capyfs

import "github.com/spf13/afero"

var Filesystem afero.Fs
var FilesystemUtils afero.Afero

func InitOsFilesystem() {
	Filesystem = afero.NewOsFs()
	FilesystemUtils = afero.Afero{Fs: Filesystem}
}

// InitCopyOnWriteFilesystem Read only os filesystem with writable mem filesystem on top.
func InitCopyOnWriteFilesystem() {
	filesystem := afero.NewCopyOnWriteFs(
		afero.NewReadOnlyFs(afero.NewOsFs()),
		afero.NewMemMapFs(),
	)

	Filesystem = filesystem
	FilesystemUtils = afero.Afero{Fs: filesystem}
}

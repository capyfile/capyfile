package capyfs

import "github.com/spf13/afero"

var OsFilesystem = afero.NewOsFs()
var OsFilesystemUtils = afero.Afero{Fs: OsFilesystem}

var MemFilesystem = afero.NewMemMapFs()

// CopyOnWriteFilesystem Read only os filesystem with writable mem filesystem on top.
var CopyOnWriteFilesystem = afero.NewCopyOnWriteFs(OsFilesystem, MemFilesystem)

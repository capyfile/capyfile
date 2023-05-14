package operations

import "capyfile/files"

type OperationHandler interface {
	Handle(in []files.ProcessableFile) (out []files.ProcessableFile, err error)
}

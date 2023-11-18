package opfactories

import "capyfile/operations"

func NewFilesystemInputRemoveOperation(name string) (*operations.FilesystemInputRemoveOperation, error) {
	return &operations.FilesystemInputRemoveOperation{
		Name:   name,
		Params: &operations.FilesystemInputRemoveOperationParams{},
	}, nil
}

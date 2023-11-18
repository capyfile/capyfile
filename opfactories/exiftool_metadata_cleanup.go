package opfactories

import "capyfile/operations"

func NewExiftoolMetadataCleanupOperation(name string) (*operations.ExiftoolMetadataCleanupOperation, error) {
	return &operations.ExiftoolMetadataCleanupOperation{
		Name: name,
	}, nil
}

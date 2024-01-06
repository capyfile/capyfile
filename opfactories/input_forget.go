package opfactories

import (
	"capyfile/operations"
)

func NewInputForgetOperation(name string) (*operations.InputForgetOperation, error) {
	return &operations.InputForgetOperation{
		Name: name,
	}, nil
}

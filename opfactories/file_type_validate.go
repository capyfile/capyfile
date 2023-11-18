package opfactories

import (
	"capyfile/operations"
	"capyfile/parameters"
	"errors"
)

func NewFileTypeValidateOperation(
	name string,
	params map[string]parameters.Parameter,
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FileTypeValidateOperation, error) {
	var allowedMimeTypes []string
	if allowedMimeTypeParameter, ok := params["allowedMimeTypes"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			allowedMimeTypeParameter.SourceType,
			allowedMimeTypeParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringArrayValue()
		if valErr != nil {
			return nil, valErr
		}

		allowedMimeTypes = val
	} else {
		return nil, errors.New("failed to retrieve \"allowedMimeTypes\" parameter")
	}

	return &operations.FileTypeValidateOperation{
		Name: name,
		Params: &operations.FileTypeValidateOperationParams{
			AllowedMimeTypes: allowedMimeTypes,
		},
	}, nil
}

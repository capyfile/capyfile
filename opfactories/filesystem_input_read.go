package opfactories

import (
	"capyfile/operations"
	"capyfile/parameters"
	"errors"
)

func NewFilesystemInputReadOperation(
	name string,
	params map[string]parameters.Parameter,
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FilesystemInputReadOperation, error) {
	var target string
	if targetParameter, ok := params["target"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			targetParameter.SourceType,
			targetParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		target = val
	} else {
		return nil, errors.New("failed to retrieve \"target\" parameter")
	}

	return &operations.FilesystemInputReadOperation{
		Name: name,
		Params: &operations.FilesystemInputReadOperationParams{
			Target: target,
		},
	}, nil
}

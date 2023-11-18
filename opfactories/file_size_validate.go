package opfactories

import (
	"capyfile/operations"
	"capyfile/parameters"
	"errors"
)

func NewFileSizeValidateOperation(
	name string,
	params map[string]parameters.Parameter,
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FileSizeValidateOperation, error) {
	var minFileSize int64 = 0
	if minFileSizeParameter, ok := params["minFileSize"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			minFileSizeParameter.SourceType,
			minFileSizeParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadIntValue()
		if valErr != nil {
			return nil, valErr
		}

		minFileSize = val
	}

	var maxFileSize int64 = 0
	if maxFileSizeParameter, ok := params["maxFileSize"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			maxFileSizeParameter.SourceType,
			maxFileSizeParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadIntValue()
		if valErr != nil {
			return nil, valErr
		}

		maxFileSize = val
	}

	if minFileSize == 0 && maxFileSize == 0 {
		return nil, errors.New("either \"minFileSize\" or \"maxFileSize\" parameter must be set")
	}

	return &operations.FileSizeValidateOperation{
		Name: name,
		Params: &operations.FileSizeValidateOperationParams{
			MinFileSize: minFileSize,
			MaxFileSize: maxFileSize,
		},
	}, nil
}

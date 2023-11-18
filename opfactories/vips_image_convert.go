package opfactories

import (
	"capyfile/operations"
	"capyfile/parameters"
	"errors"
)

func NewImageConvertOperation(
	name string,
	params map[string]parameters.Parameter,
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.ImageConvertOperation, error) {
	var toMimeType string
	if toMimeTypeParameter, ok := params["toMimeType"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			toMimeTypeParameter.SourceType,
			toMimeTypeParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		toMimeType = val
	} else {
		return nil, errors.New("failed to retrieve \"toMimeType\" parameter")
	}

	var quality string
	if toMimeTypeParameter, ok := params["quality"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			toMimeTypeParameter.SourceType,
			toMimeTypeParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		quality = val
	} else {
		return nil, errors.New("failed to retrieve \"quality\" parameter")
	}

	return &operations.ImageConvertOperation{
		Name: name,
		Params: &operations.ImageConvertOperationParams{
			ToMimeType: toMimeType,
			Quality:    quality,
		},
	}, nil
}

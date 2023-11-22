package opfactories

import (
	"capyfile/operations"
	"capyfile/parameters"
)

func NewExiftoolMetadataCleanupOperation(
	name string,
	params map[string]parameters.Parameter,
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.ExiftoolMetadataCleanupOperation, error) {
	var overrideOriginalFile bool
	if destinationParameter, ok := params["overrideOriginalFile"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			destinationParameter.SourceType,
			destinationParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadBoolValue()
		if valErr != nil {
			return nil, valErr
		}

		overrideOriginalFile = val
	}

	return &operations.ExiftoolMetadataCleanupOperation{
		Name: name,
		Params: &operations.ExiftoolMetadataCleanupOperationParams{
			OverwriteOriginalFile: overrideOriginalFile,
		},
	}, nil
}

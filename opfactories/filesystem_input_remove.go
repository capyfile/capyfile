package opfactories

import (
	"capyfile/operations"
	"capyfile/parameters"
	"errors"
)

func NewFilesystemInputRemoveOperation(
	name string,
	params map[string]parameters.Parameter,
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FilesystemInputRemoveOperation, error) {
	var removeOriginalFile bool
	if destinationParameter, ok := params["removeOriginalFile"]; ok {
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

		removeOriginalFile = val
	} else {
		return nil, errors.New("failed to retrieve \"removeOriginalFile\" parameter")
	}

	return &operations.FilesystemInputRemoveOperation{
		Name: name,
		Params: &operations.FilesystemInputRemoveOperationParams{
			RemoveOriginalFile: removeOriginalFile,
		},
	}, nil
}

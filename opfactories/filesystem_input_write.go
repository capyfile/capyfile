package opfactories

import (
	"capyfile/operations"
	"capyfile/parameters"
	"errors"
)

func NewFilesystemInputWriteOperation(
	name string,
	params map[string]parameters.Parameter,
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FilesystemInputWriteOperation, error) {
	var destination string
	if destinationParameter, ok := params["destination"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			destinationParameter.SourceType,
			destinationParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		destination = val
	} else {
		return nil, errors.New("failed to retrieve \"destination\" parameter")
	}

	var useOriginalFilename bool
	if destinationParameter, ok := params["useOriginalFilename"]; ok {
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

		useOriginalFilename = val
	} else {
		return nil, errors.New("failed to retrieve \"useOriginalFilename\" parameter")
	}

	return &operations.FilesystemInputWriteOperation{
		Name: name,
		Params: &operations.FilesystemInputWriteOperationParams{
			Destination:         destination,
			UseOriginalFilename: useOriginalFilename,
		},
	}, nil
}

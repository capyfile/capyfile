package opfactories

import (
	"capyfile/operations"
	"capyfile/parameters"
	"errors"
)

func NewCommandExecOperation(
	name string,
	params map[string]parameters.Parameter,
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.CommandExecOperation, error) {
	var commandName string
	if commandNameParameter, ok := params["commandName"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			commandNameParameter.SourceType,
			commandNameParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		commandName = val
	} else {
		return nil, errors.New("failed to retrieve \"commandName\" parameter")
	}

	var commandArgs []string
	if commandArgsParameter, ok := params["commandArgs"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			commandArgsParameter.SourceType,
			commandArgsParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringArrayValue()
		if valErr != nil {
			return nil, valErr
		}

		commandArgs = val
	}

	var outputFileDestination string
	if outputFileDestinationParameter, ok := params["outputFileDestination"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			outputFileDestinationParameter.SourceType,
			outputFileDestinationParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		outputFileDestination = val
	}

	var allowParallelExecution bool = false
	if allowParallelExecutionParameter, ok := params["allowParallelExecution"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			allowParallelExecutionParameter.SourceType,
			allowParallelExecutionParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadBoolValue()
		if valErr != nil {
			return nil, valErr
		}

		allowParallelExecution = val
	}

	return &operations.CommandExecOperation{
		Name: name,
		Params: &operations.CommandExecOperationParams{
			CommandName:            commandName,
			CommandArgs:            commandArgs,
			OutputFileDestination:  outputFileDestination,
			AllowParallelExecution: allowParallelExecution,
		},
	}, nil
}

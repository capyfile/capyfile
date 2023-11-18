package opfactories

import (
	"capyfile/operations"
	"capyfile/operations/filetime"
	"capyfile/parameters"
	"errors"
	"time"
)

func NewFileTimeValidateOperation(
	name string,
	params map[string]parameters.Parameter,
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FileTimeValidateOperation, error) {
	timeParamValExtractor := func(paramName string) (time.Time, error) {
		var paramVal time.Time
		if param, ok := params[paramName]; ok {
			parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
				param.SourceType,
				param.Source,
			)
			if loaderErr != nil {
				return paramVal, loaderErr
			}

			val, valErr := parameterLoader.LoadStringValue()
			if valErr != nil {
				return paramVal, valErr
			}

			timeVal, timeParseErr := time.Parse(time.RFC3339, val)
			if timeParseErr != nil {
				return paramVal, timeParseErr
			}

			paramVal = timeVal
		}

		return paramVal, nil
	}

	minAtime, minAtimeErr := timeParamValExtractor("minAtime")
	if minAtimeErr != nil {
		return nil, minAtimeErr
	}

	maxAtime, maxAtimeErr := timeParamValExtractor("maxAtime")
	if maxAtimeErr != nil {
		return nil, maxAtimeErr
	}

	minMtime, minMtimeErr := timeParamValExtractor("minMtime")
	if minMtimeErr != nil {
		return nil, minMtimeErr
	}

	maxMtime, maxMtimeErr := timeParamValExtractor("maxMtime")
	if maxMtimeErr != nil {
		return nil, maxMtimeErr
	}

	minCtime, minCtimeErr := timeParamValExtractor("minCtime")
	if minCtimeErr != nil {
		return nil, minCtimeErr
	}

	maxCtime, maxCtimeErr := timeParamValExtractor("maxCtime")
	if maxCtimeErr != nil {
		return nil, maxCtimeErr
	}

	if minAtime.IsZero() &&
		maxAtime.IsZero() &&
		minMtime.IsZero() &&
		maxMtime.IsZero() &&
		minCtime.IsZero() &&
		maxCtime.IsZero() {
		return nil, errors.New(
			"either \"minAtime\", \"maxAtime\", \"minMtime\", \"maxMtime\", " +
				"\"minCtime\", or \"maxCtime\" parameter must be set",
		)
	}

	return &operations.FileTimeValidateOperation{
		Name: name,
		Params: &operations.FileTimeValidateOperationParams{
			MinAtime: minAtime,
			MaxAtime: maxAtime,
			MinMtime: minMtime,
			MaxMtime: maxMtime,
			MinCtime: minCtime,
			MaxCtime: maxCtime,
		},
		TimeStatProvider: &filetime.PlatformTimeStatProvider{},
	}, nil
}

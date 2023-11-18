package opfactories

import (
	"capyfile/operations"
	"capyfile/parameters"
	"errors"
)

func NewS3UploadOperation(
	name string,
	params map[string]parameters.Parameter,
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.S3UploadOperation, error) {
	var accessKeyId = ""
	if accessKeyIdParameter, ok := params["accessKeyId"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			accessKeyIdParameter.SourceType,
			accessKeyIdParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		accessKeyId = val
	} else {
		return nil, errors.New("failed to retrieve \"accessKeyId\" parameter")
	}

	var secretAccessKey = ""
	if secretAccessKeyParameter, ok := params["secretAccessKey"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			secretAccessKeyParameter.SourceType,
			secretAccessKeyParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		secretAccessKey = val
	} else {
		return nil, errors.New("failed to retrieve \"secretAccessKey\" parameter")
	}

	var sessionToken = ""
	if sessionTokenParameter, ok := params["sessionToken"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			sessionTokenParameter.SourceType,
			sessionTokenParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		sessionToken = val
	}

	var endpoint = ""
	if endpointParameter, ok := params["endpoint"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			endpointParameter.SourceType,
			endpointParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		endpoint = val
	} else {
		return nil, errors.New("failed to retrieve \"endpoint\" parameter")
	}

	var region = ""
	if regionParameter, ok := params["region"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			regionParameter.SourceType,
			regionParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		region = val
	} else {
		return nil, errors.New("failed to retrieve \"region\" parameter")
	}

	var bucket = ""
	if bucketParameter, ok := params["bucket"]; ok {
		parameterLoader, loaderErr := parameterLoaderProvider.ParameterLoader(
			bucketParameter.SourceType,
			bucketParameter.Source,
		)
		if loaderErr != nil {
			return nil, loaderErr
		}

		val, valErr := parameterLoader.LoadStringValue()
		if valErr != nil {
			return nil, valErr
		}

		bucket = val
	} else {
		return nil, errors.New("failed to retrieve \"bucket\" parameter")
	}

	return &operations.S3UploadOperation{
		Name: name,
		Params: &operations.S3UploadOperationParams{
			AccessKeyId:     accessKeyId,
			SecretAccessKey: secretAccessKey,
			SessionToken:    sessionToken,
			Endpoint:        endpoint,
			Region:          region,
			Bucket:          bucket,
		},
	}, nil
}

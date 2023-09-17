package capysvc

import (
	"capyfile/capyerr"
	"capyfile/capysvc/processor"
	"capyfile/capysvc/service"
	"capyfile/files"
	"capyfile/operations"
	"capyfile/parameters"
	"errors"
	"fmt"
)

type Service struct {
	Version    string      `json:"version"`
	Name       string      `json:"name"`
	Processors []Processor `json:"processors"`
}

func (s *Service) FindProcessor(processorName string) *Processor {
	for _, p := range s.Processors {
		if p.Name == processorName {
			return &p
		}
	}

	return nil
}

func (s *Service) RunProcessor(
	serviceContext service.Context,
	processorName string,
	in []files.ProcessableFile,
) (out []files.ProcessableFile, err error) {
	proc := s.FindProcessor(processorName)
	if proc == nil {
		return out, capyerr.NewProcessorNotFoundError(processorName)
	}

	return proc.RunOperations(serviceContext.ProcessorContext(), in)
}

type Processor struct {
	Name       string      `json:"name"`
	Operations []Operation `json:"operations"`
}

func (p *Processor) RunOperations(
	processorContext processor.Context,
	in []files.ProcessableFile,
) (out []files.ProcessableFile, err error) {
	out = in

	parameterLoaderProvider, providerErr := processorContext.ParameterLoaderProvider()
	if providerErr != nil {
		return out, providerErr
	}

	for _, op := range p.Operations {
		opHandler, opHandlerError := op.OperationHandler(parameterLoaderProvider)
		if opHandlerError != nil {
			return out, opHandlerError
		}

		out, err = opHandler.Handle(out)
		if err != nil {
			return out, err
		}
	}

	return out, nil
}

type Operation struct {
	Name   string                        `json:"name"`
	Params map[string]OperationParameter `json:"params"`
}

// OperationHandler Factory method that builds operation handler from operation definition.
func (o *Operation) OperationHandler(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (operations.OperationHandler, error) {
	switch o.Name {
	case "file_size_validate":
		return o.newFileSizeValidateOperation(parameterLoaderProvider)
	case "file_type_validate":
		return o.newFileTypeValidateOperation(parameterLoaderProvider)
	case "metadata_cleanup":
		return o.newMetadataCleanupOperation(parameterLoaderProvider)
	case "image_convert":
		return o.newImageConvertOperation(parameterLoaderProvider)
	case "s3_upload":
		return o.newS3UploadOperation(parameterLoaderProvider)
	case "s3_upload_v2":
		return o.newS3UploadV2Operation(parameterLoaderProvider)
	case "filesystem_input_read":
		return o.newFilesystemInputReadOperation(parameterLoaderProvider)
	case "filesystem_input_write":
		return o.newFilesystemInputWriteOperation(parameterLoaderProvider)
	default:
		return nil, fmt.Errorf("unknown operation \"%s\"", o.Name)
	}
}

type OperationParameter struct {
	SourceType string `json:"sourceType"`
	Source     any    `json:"source"`
}

func (o *Operation) newFileSizeValidateOperation(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FileSizeValidateOperation, error) {
	var minFileSize int64 = 0
	if minFileSizeParameter, ok := o.Params["minFileSize"]; ok {
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
	} else {
		return nil, errors.New("failed to retrieve \"minFileSize\" parameter")
	}

	var maxFileSize int64 = 0
	if maxFileSizeParameter, ok := o.Params["maxFileSize"]; ok {
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
	} else {
		return nil, errors.New("failed to retrieve \"maxFileSize\" parameter")
	}

	return &operations.FileSizeValidateOperation{
		Params: &operations.FileSizeValidateOperationParams{
			MinFileSize: minFileSize,
			MaxFileSize: maxFileSize,
		},
	}, nil
}

func (o *Operation) newFileTypeValidateOperation(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FileTypeValidateOperation, error) {
	var allowedMimeTypes []string
	if allowedMimeTypeParameter, ok := o.Params["allowedMimeTypes"]; ok {
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
		Params: &operations.FileTypeValidateOperationParams{
			AllowedMimeTypes: allowedMimeTypes,
		},
	}, nil
}

func (o *Operation) newMetadataCleanupOperation(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.MetadataCleanupOperation, error) {
	return &operations.MetadataCleanupOperation{}, nil
}

func (o *Operation) newImageConvertOperation(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.ImageConvertOperation, error) {
	var toMimeType string
	if toMimeTypeParameter, ok := o.Params["toMimeType"]; ok {
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
	if toMimeTypeParameter, ok := o.Params["quality"]; ok {
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
		Params: &operations.ImageConvertOperationParams{
			ToMimeType: toMimeType,
			Quality:    quality,
		},
	}, nil
}

func (o *Operation) newS3UploadOperation(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.S3UploadOperation, error) {
	var accessKeyId = ""
	if accessKeyIdParameter, ok := o.Params["accessKeyId"]; ok {
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
	if secretAccessKeyParameter, ok := o.Params["secretAccessKey"]; ok {
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
	if sessionTokenParameter, ok := o.Params["sessionToken"]; ok {
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
	if endpointParameter, ok := o.Params["endpoint"]; ok {
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
	if regionParameter, ok := o.Params["region"]; ok {
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
	if bucketParameter, ok := o.Params["bucket"]; ok {
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

func (o *Operation) newS3UploadV2Operation(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.S3UploadV2Operation, error) {
	var accessKeyId = ""
	if accessKeyIdParameter, ok := o.Params["accessKeyId"]; ok {
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
	if secretAccessKeyParameter, ok := o.Params["secretAccessKey"]; ok {
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
	if sessionTokenParameter, ok := o.Params["sessionToken"]; ok {
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
	if endpointParameter, ok := o.Params["endpoint"]; ok {
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
	if regionParameter, ok := o.Params["region"]; ok {
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
	if bucketParameter, ok := o.Params["bucket"]; ok {
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

	return &operations.S3UploadV2Operation{
		Params: &operations.S3UploadV2OperationParams{
			AccessKeyId:     accessKeyId,
			SecretAccessKey: secretAccessKey,
			SessionToken:    sessionToken,
			Endpoint:        endpoint,
			Region:          region,
			Bucket:          bucket,
		},
	}, nil
}

func (o *Operation) newFilesystemInputReadOperation(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FilesystemInputReadOperation, error) {
	var target string
	if targetParameter, ok := o.Params["target"]; ok {
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
		Params: &operations.FilesystemInputReadOperationParams{
			Target: target,
		},
	}, nil
}

func (o *Operation) newFilesystemInputWriteOperation(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FilesystemInputWriteOperation, error) {
	var destination string
	if destinationParameter, ok := o.Params["destination"]; ok {
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
	if destinationParameter, ok := o.Params["useOriginalFilename"]; ok {
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
		Params: &operations.FilesystemInputWriteOperationParams{
			Destination:         destination,
			UseOriginalFilename: useOriginalFilename,
		},
	}, nil
}

package capysvc

import (
	"capyfile/capyerr"
	"capyfile/files"
	"capyfile/operations"
	"capyfile/operations/filetime"
	"capyfile/parameters"
	"errors"
	"fmt"
	"sync"
	"time"
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
	context Context,
	processorName string,
	in []files.ProcessableFile,
) (out []files.ProcessableFile, err error) {
	proc := s.FindProcessor(processorName)
	if proc == nil {
		return out, capyerr.NewProcessorNotFoundError(processorName)
	}

	return proc.RunOperations(context, in)
}

func (s *Service) RunProcessorConcurrently(
	ctx Context,
	processorName string,
	in []files.ProcessableFile,
	errorCh chan<- operations.OperationError,
	notificationCh chan<- operations.OperationNotification,
) (out []files.ProcessableFile, err error) {
	proc := s.FindProcessor(processorName)
	if proc == nil {
		return out, capyerr.NewProcessorNotFoundError(processorName)
	}

	return proc.RunOperationsConcurrently(
		ctx,
		in,
		errorCh,
		notificationCh,
	)
}

type Processor struct {
	Name       string      `json:"name"`
	Operations []Operation `json:"operations"`
}

func (p *Processor) RunOperations(
	ctx Context,
	in []files.ProcessableFile,
) (out []files.ProcessableFile, err error) {
	iniOpsErr := p.initOperations(ctx)
	if iniOpsErr != nil {
		return out, iniOpsErr
	}

	firstOp, _ := p.firstAndLastOperations()
	firstOp.enqueueIn(in...)
	firstOp.incrInCount(len(in))

	for i := range p.Operations {
		op := &p.Operations[i]

		out, err = op.handler.Handle(op.dequeueAllIn(), nil, nil)
		if err != nil {
			return out, err
		}

		if op.nextOperation != nil {
			op.nextOperation.enqueueIn(out...)
		}
	}

	return out, nil
}

func (p *Processor) RunOperationsConcurrently(
	ctx Context,
	in []files.ProcessableFile,
	errorCh chan<- operations.OperationError,
	notificationCh chan<- operations.OperationNotification,
) (out []files.ProcessableFile, err error) {
	iniOpsErr := p.initOperations(ctx)
	if iniOpsErr != nil {
		return out, iniOpsErr
	}

	firstOp, lastOp := p.firstAndLastOperations()

	firstOp.enqueueIn(in...)
	firstOp.incrInCount(len(in))

	outCh := make(chan []files.ProcessableFile)

	for i := range p.Operations {
		op := &p.Operations[i]

		go func(op *Operation) {
			if op.handler.AllowConcurrency() {
				for !op.Completed {
					opIn := op.dequeueIn(1)
					if len(opIn) == 0 {
						continue
					}

					go op.handleAndPassOutput(opIn, outCh, errorCh, notificationCh)
				}
			} else {
				if op.prevOperation != nil {
					for !op.prevOperation.Completed {
						// Wait for all the input for the previous operations to be processed.
					}
				}

				opIn := op.dequeueAllIn()
				op.handleAndPassOutput(opIn, outCh, errorCh, notificationCh)
			}
		}(op)
	}

	go func() {
		for !lastOp.Completed {
			// Wait for the last operation to be completed.
		}
		close(outCh)
	}()

	for o := range outCh {
		out = append(out, o...)
	}

	return out, nil
}

// initOperations initializes the operations before running the pipeline.
//
// What this method does is:
//   - initializes the operation handlers
//   - builds the linked list of operations
func (p *Processor) initOperations(ctx Context) error {
	var prevOp *Operation
	for i := range p.Operations {
		op := &p.Operations[i]

		// If the target files is not set, we set it to "without_errors" by default so
		// the handler will only handle the input without errors.
		if op.TargetFiles == "" {
			op.TargetFiles = OperationTargetFilesWithoutErrors
		}

		op.handlerLock = &sync.Mutex{}
		op.inLock = &sync.Mutex{}
		op.inCntLock = &sync.Mutex{}
		op.completedLock = &sync.Mutex{}

		opHandlerErr := op.initOperationHandler(ctx)
		if opHandlerErr != nil {
			return opHandlerErr
		}

		if prevOp != nil {
			prevOp.nextOperation = op
			prevOp.nextOperation.prevOperation = prevOp
		}

		prevOp = op
	}

	return nil
}

func (p *Processor) firstAndLastOperations() (firstOp *Operation, lastOp *Operation) {
	if len(p.Operations) == 0 {
		return nil, nil
	}

	return &p.Operations[0], &p.Operations[len(p.Operations)-1]
}

const (
	OperationTargetFilesAll           = "all"
	OperationTargetFilesWithoutErrors = "without_errors"
	OperationTargetFilesWithErrors    = "with_errors"
)

type Operation struct {
	Name        string                        `json:"name"`
	TargetFiles string                        `json:"targetFiles"`
	Params      map[string]OperationParameter `json:"params"`

	inLock *sync.Mutex
	in     []files.ProcessableFile
	// We need this counter to know when the operation is completed.
	// Queueing the input is increasing the counter, when the input is processed
	// we are decreasing the counter. But we can't rely on this when we deal with
	// the empty input.
	inCntLock *sync.Mutex
	InCnt     int

	completedLock *sync.Mutex
	// Whether the operation is completed which means that there is no more input
	// for the operation and all the input is processed.
	Completed bool

	// Here we are holding the initialized operation handlers.
	// Should be modified if you want to reload parameters more often.
	handlerLock *sync.Mutex
	handler     operations.OperationHandler

	prevOperation *Operation
	nextOperation *Operation
}

// handleAndPassOutput does the actual input handling and passing the output to
// the next operation or to the final output channel.
//
// Keep in mind that outCh returns the actual final output for the pipeline,
// not the individual operation output. The individual operation output is
// passed to the next operation inside this method.
func (o *Operation) handleAndPassOutput(
	in []files.ProcessableFile,
	outCh chan<- []files.ProcessableFile,
	errorCh chan<- operations.OperationError,
	notificationCh chan<- operations.OperationNotification,
) {
	var targetIn []files.ProcessableFile
	var skipIn []files.ProcessableFile

	if o.TargetFiles != OperationTargetFilesAll {
		if o.TargetFiles == OperationTargetFilesWithoutErrors {
			for _, pf := range in {
				if !pf.HasFileProcessingError() {
					targetIn = append(targetIn, pf)
				} else {
					skipIn = append(skipIn, pf)
				}
			}
		} else if o.TargetFiles == OperationTargetFilesWithErrors {
			for _, pf := range in {
				if pf.HasFileProcessingError() {
					targetIn = append(targetIn, pf)
				} else {
					skipIn = append(skipIn, pf)
				}
			}
		}

		// We can pass the skipped input to the next operation if there is any.
		if len(skipIn) > 0 {
			for _, pf := range skipIn {
				if notificationCh != nil {
					notificationCh <- operations.NewOperationNotification(
						o.Name,
						operations.StatusSkipped,
						fmt.Sprintf("skipped due to \"%s\" target files policy", o.TargetFiles),
						&pf,
						nil,
					)
				}
			}

			if o.nextOperation != nil {
				// If we have the next operation, we need to enqueue the skipped input to it.
				o.nextOperation.incrInCount(len(skipIn))
				o.nextOperation.enqueueIn(skipIn...)
			} else {
				// This is the last operation, so we can just add the skipped input to the final output.
				outCh <- skipIn
			}

			o.decrInCount(len(in))
		}
	} else {
		targetIn = in
	}

	opHandlerOut, opHandlerErr := o.handler.Handle(targetIn, errorCh, notificationCh)
	if opHandlerErr != nil {
		// Perhaps maybe this is something that is related to the specific file,
		// so we can just send an error to the channel and continue.
		if errorCh != nil {
			errorCh <- operations.NewOperationInputError(o.Name, targetIn, opHandlerErr)
		}
	}

	if o.nextOperation != nil {
		// If we have the next operation, we need to enqueue the output to it.
		o.nextOperation.incrInCount(len(opHandlerOut))
		o.nextOperation.enqueueIn(opHandlerOut...)
	} else {
		// This is the last operation, so we can just add the output to the final output.
		outCh <- opHandlerOut
	}

	o.decrInCount(len(targetIn))

	if !o.Completed {
		if o.prevOperation != nil {
			// Here we check the previous operation. If the previous operation is completed
			// and there is no input for the current operation, we can say that the current
			// operation is completed.
			if o.prevOperation.Completed && o.InCnt == 0 {
				o.complete()
			}
		} else {
			// If this is the first operation, we can say that it's completed if there is
			// no input for it.
			if o.InCnt == 0 {
				o.complete()
			}
		}
	}
}

func (o *Operation) complete() {
	o.completedLock.Lock()
	defer o.completedLock.Unlock()

	o.Completed = true
}

func (o *Operation) sizeIn() int {
	o.inLock.Lock()
	defer o.inLock.Unlock()

	return len(o.in)
}

func (o *Operation) enqueueIn(pf ...files.ProcessableFile) {
	o.inLock.Lock()
	defer o.inLock.Unlock()

	o.in = append(o.in, pf...)
}

func (o *Operation) dequeueIn(size int) []files.ProcessableFile {
	o.inLock.Lock()
	defer o.inLock.Unlock()

	if len(o.in) == 0 {
		return nil
	}

	var dequeuedIn []files.ProcessableFile
	if size > len(o.in) {
		dequeuedIn = o.in
		o.in = nil
	} else {
		dequeuedIn = o.in[:size]
		o.in = o.in[size:]
	}

	return dequeuedIn
}

func (o *Operation) dequeueAllIn() []files.ProcessableFile {
	o.inLock.Lock()
	defer o.inLock.Unlock()

	in := o.in
	o.in = nil

	return in
}

func (o *Operation) incrInCount(i int) {
	o.inCntLock.Lock()
	defer o.inCntLock.Unlock()

	o.InCnt += i
}

func (o *Operation) decrInCount(i int) {
	o.inCntLock.Lock()
	defer o.inCntLock.Unlock()

	o.InCnt -= i
}

func (o *Operation) initOperationHandler(ctx Context) error {
	// If the handler is already created, return it.
	if o.handler != nil {
		return nil
	}

	parameterLoaderProvider, providerErr := ctx.ParameterLoaderProvider()
	if providerErr != nil {
		return providerErr
	}

	o.handlerLock.Lock()
	defer o.handlerLock.Unlock()

	if o.handler != nil {
		return nil
	}

	var oh operations.OperationHandler
	var ohErr error

	switch o.Name {
	case "file_size_validate":
		oh, ohErr = o.newFileSizeValidateOperation(parameterLoaderProvider)
		break
	case "file_type_validate":
		oh, ohErr = o.newFileTypeValidateOperation(parameterLoaderProvider)
		break
	case "file_time_validate":
		oh, ohErr = o.newFileTimeValidateOperation(parameterLoaderProvider)
		break
	case "exiftool_metadata_cleanup":
		oh, ohErr = o.newExiftoolMetadataCleanupOperation(parameterLoaderProvider)
		break
	case "image_convert":
		oh, ohErr = o.newImageConvertOperation(parameterLoaderProvider)
		break
	case "s3_upload":
		oh, ohErr = o.newS3UploadOperation(parameterLoaderProvider)
		break
	//case "s3_upload_v2":
	//	oh, ohErr = o.newS3UploadV2Operation(parameterLoaderProvider)
	//	break
	case "filesystem_input_read":
		oh, ohErr = o.newFilesystemInputReadOperation(parameterLoaderProvider)
		break
	case "filesystem_input_write":
		oh, ohErr = o.newFilesystemInputWriteOperation(parameterLoaderProvider)
		break
	case "filesystem_input_remove":
		oh, ohErr = o.newFilesystemInputRemoveOperation(parameterLoaderProvider)
		break
	default:
		return fmt.Errorf("unknown operation \"%s\"", o.Name)
	}

	if ohErr == nil {
		o.handler = oh

		return nil
	}

	return ohErr
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
	}

	if minFileSize == 0 && maxFileSize == 0 {
		return nil, errors.New("either \"minFileSize\" or \"maxFileSize\" parameter must be set")
	}

	return &operations.FileSizeValidateOperation{
		Name: o.Name,
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
		Name: o.Name,
		Params: &operations.FileTypeValidateOperationParams{
			AllowedMimeTypes: allowedMimeTypes,
		},
	}, nil
}

func (o *Operation) newFileTimeValidateOperation(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FileTimeValidateOperation, error) {
	timeParamValExtractor := func(paramName string) (time.Time, error) {
		var paramVal time.Time
		if param, ok := o.Params[paramName]; ok {
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
		Name: o.Name,
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

func (o *Operation) newExiftoolMetadataCleanupOperation(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.ExiftoolMetadataCleanupOperation, error) {
	return &operations.ExiftoolMetadataCleanupOperation{
		Name: o.Name,
	}, nil
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
		Name: o.Name,
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
		Name: o.Name,
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
		Name: o.Name,
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
		Name: o.Name,
		Params: &operations.FilesystemInputWriteOperationParams{
			Destination:         destination,
			UseOriginalFilename: useOriginalFilename,
		},
	}, nil
}

func (o *Operation) newFilesystemInputRemoveOperation(
	parameterLoaderProvider parameters.ParameterLoaderProvider,
) (*operations.FilesystemInputRemoveOperation, error) {
	return &operations.FilesystemInputRemoveOperation{
		Name:   o.Name,
		Params: &operations.FilesystemInputRemoveOperationParams{},
	}, nil
}

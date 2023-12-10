package capysvc

import (
	"capyfile/capyerr"
	"capyfile/files"
	"capyfile/operations"
	"capyfile/opfactories"
	"capyfile/parameters"
	"errors"
	"fmt"
	"sync"
)

type Service struct {
	Version    string      `json:"version" yaml:"version"`
	Name       string      `json:"name" yaml:"name"`
	Processors []Processor `json:"processors" yaml:"processors"`
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
	ctx Context,
	processorName string,
	in []files.ProcessableFile,
) (out []files.ProcessableFile, err error) {
	proc := s.FindProcessor(processorName)
	if proc == nil {
		return out, capyerr.NewProcessorNotFoundError(processorName)
	}

	iniOpsErr := proc.InitOperations(ctx)
	if iniOpsErr != nil {
		return out, iniOpsErr
	}

	return proc.RunOperations(ctx, in)
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

	iniOpsErr := proc.InitOperations(ctx)
	if iniOpsErr != nil {
		return out, iniOpsErr
	}

	return proc.RunOperationsConcurrently(
		ctx,
		in,
		errorCh,
		notificationCh,
	)
}

type Processor struct {
	Name       string      `json:"name" yaml:"name"`
	Operations []Operation `json:"operations" yaml:"operations"`
}

func (p *Processor) RunOperations(
	ctx Context,
	in []files.ProcessableFile,
) (out []files.ProcessableFile, err error) {
	firstOp, _ := p.firstAndLastOperations()
	firstOp.enqueueIn(in...)
	firstOp.incrInCnt(len(in))

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
	firstOp, lastOp := p.firstAndLastOperations()

	firstOp.enqueueIn(in...)
	firstOp.incrInCnt(len(in))

	outCh := make(chan []files.ProcessableFile)

	for i := range p.Operations {
		op := &p.Operations[i]

		go func(op *Operation) {
			if op.handler.AllowConcurrency() {
				for !op.Completed {
					if op.prevOperation != nil {
						// If what we deal with here is not the first operation, we need to check
						// whether the previous operation has produced any output. If not, then
						// the current concurrent operation can be skipped.
						if op.prevOperation.Completed && op.prevOperation.OutCnt() == 0 {
							op.complete()
							break
						}
					} else {
						// If the first operation is happened to be concurrent, we need to check
						// whether there is any input for it. If not, then the current concurrent
						// operation can be skipped.
						if op.InCnt() == 0 {
							op.complete()
							break
						}
					}

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

// InitOperations initializes the operations before running the pipeline.
//
// Use this method when you run the pipeline first time or when you want to
// reconfigure the pipeline. But keep in mind that this method is not safe
// run on the pipeline that is running. You need to stop the pipeline first,
// or wait until it's completed.
//
// What this method does is:
//   - resets the operations to their initial state
//   - initializes the operation handlers
//   - builds the linked list of operations
func (p *Processor) InitOperations(ctx Context) error {
	p.resetOperations()

	var prevOp *Operation
	for i := range p.Operations {
		op := &p.Operations[i]

		// If the target files is not set, we set it to "without_errors" by default so
		// the handler will only handle the input without errors.
		if op.TargetFiles == "" {
			op.TargetFiles = OperationTargetFilesWithoutErrors
		}
		// If the cleanup policy is not set, we set it to "keep_files" by default so
		// the handler will keep the files produced by the operation.
		if op.CleanupPolicy == "" {
			op.CleanupPolicy = OperationCleanupPolicyKeepFiles
		}

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

func (p *Processor) resetOperations() {
	for i := range p.Operations {
		op := &p.Operations[i]

		op.inLock = &sync.Mutex{}
		op.in = []files.ProcessableFile{}

		op.inOutCntLock = &sync.Mutex{}
		op.inCnt = 0
		op.outCnt = 0

		op.completedLock = &sync.Mutex{}
		op.Completed = false

		op.handlerLock = &sync.Mutex{}
		op.handler = nil

		op.prevOperation = nil
		op.nextOperation = nil
	}
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

const (
	OperationCleanupPolicyKeepFiles   = "keep_files"
	OperationCleanupPolicyRemoveFiles = "remove_files"
)

type Operation struct {
	Name   string                          `json:"name" yaml:"name"`
	Params map[string]parameters.Parameter `json:"params" yaml:"params"`
	// TargetFiles is the parameter that defines which files should be handled by the operation.
	// The possible values are:
	//   - without_errors - only the input files without errors (default)
	//   - with_errors - only the input files with errors
	//   - all - all the input files
	TargetFiles string `json:"targetFiles" yaml:"targetFiles"`
	// CleanupPolicy is the parameter that defines what should be done with the files that
	// this operation has produced when it's time to perform the cleanup to free the resources.
	// The possible values are:
	//   - keep_files - keep the files (default)
	//   - remove_files - remove the files
	CleanupPolicy string `json:"cleanupPolicy" yaml:"cleanupPolicy"`

	inLock *sync.Mutex
	in     []files.ProcessableFile

	inOutCntLock *sync.Mutex
	// We need inCnt to know when the operation is completed.
	// Queueing the input is increasing the counter, when the input is processed
	// we are decreasing the counter. But we can't rely on this when we deal with
	// the empty input.
	inCnt int
	// outCnt is needed for the cases when we want to know whether the operation
	// produced any output.
	// For example, if the operation is completed and there is no output, we
	// should skip any concurrent operations because by their nature they
	// can't work with empty input.
	outCnt int

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
				o.nextOperation.incrInCnt(len(skipIn))
				o.nextOperation.enqueueIn(skipIn...)
			} else {
				// This is the last operation, so we can just add the skipped input to the final output.
				outCh <- skipIn
			}

			o.decrInCntAndIncrOutCnt(len(skipIn), len(skipIn))
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

	// Here we need to set the cleanup policy for the files processed by the operation.
	// Keep in mind that the policy will actually be set for the files that does
	// not have it already set.
	for i := range opHandlerOut {
		pf := &opHandlerOut[i]
		switch o.CleanupPolicy {
		case OperationCleanupPolicyKeepFiles:
			pf.KeepOnFreeResources()
			break
		case OperationCleanupPolicyRemoveFiles:
			pf.RemoveOnFreeResources()
			break
		}
	}

	if o.nextOperation != nil {
		// If we have the next operation, we need to enqueue the output to it.
		o.nextOperation.incrInCnt(len(opHandlerOut))
		o.nextOperation.enqueueIn(opHandlerOut...)
	} else {
		// This is the last operation, so we can just add the output to the final output.
		outCh <- opHandlerOut
	}

	o.decrInCntAndIncrOutCnt(len(targetIn), len(opHandlerOut))

	if !o.Completed {
		if o.prevOperation != nil {
			// Here we check the previous operation. If the previous operation is completed
			// and there is no input for the current operation, we can say that the current
			// operation is completed.
			if o.prevOperation.Completed && o.InCnt() == 0 {
				o.complete()
			}
		} else {
			// If this is the first operation, we can say that it's completed if there is
			// no input for it.
			if o.InCnt() == 0 {
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

func (o *Operation) InCnt() int {
	o.inOutCntLock.Lock()
	defer o.inOutCntLock.Unlock()

	return o.inCnt
}

func (o *Operation) OutCnt() int {
	o.inOutCntLock.Lock()
	defer o.inOutCntLock.Unlock()

	return o.outCnt
}

func (o *Operation) decrInCntAndIncrOutCnt(iIn, iOut int) {
	o.inOutCntLock.Lock()
	defer o.inOutCntLock.Unlock()

	o.inCnt -= iIn
	o.outCnt += iOut
}

func (o *Operation) incrInCnt(i int) {
	o.inOutCntLock.Lock()
	defer o.inOutCntLock.Unlock()

	o.inCnt += i
}

func (o *Operation) initOperationHandler(ctx Context) error {
	// If the handler is already created, return it.
	if o.handler != nil {
		return nil
	}

	o.handlerLock.Lock()
	defer o.handlerLock.Unlock()

	if o.handler != nil {
		return nil
	}

	var oh operations.OperationHandler
	var ohErr error

	parameterLoaderProvider, parameterLoaderProviderErr := ctx.ParameterLoaderProvider()
	if parameterLoaderProviderErr != nil {
		return parameterLoaderProviderErr
	}

	switch o.Name {
	case "http_multipart_form_input_read":
		req := ctx.Request()
		if req == nil {
			return errors.New("http request is not available in the given context")
		}

		oh, ohErr = opfactories.NewHttpMultipartFormInputReadOperation(o.Name, ctx.Request())
		break
	case "http_octet_stream_input_read":
		req := ctx.Request()
		if req == nil {
			return errors.New("http request is not available in the given context")
		}

		oh, ohErr = opfactories.NewHttpOctetStreamInputReadOperation(o.Name, ctx.Request())
		break
	case "file_size_validate":
		oh, ohErr = opfactories.NewFileSizeValidateOperation(
			o.Name,
			o.Params,
			parameterLoaderProvider,
		)
		break
	case "file_type_validate":
		oh, ohErr = opfactories.NewFileTypeValidateOperation(
			o.Name,
			o.Params,
			parameterLoaderProvider,
		)
		break
	case "file_time_validate":
		oh, ohErr = opfactories.NewFileTimeValidateOperation(
			o.Name,
			o.Params,
			parameterLoaderProvider,
		)
		break
	case "exiftool_metadata_cleanup":
		oh, ohErr = opfactories.NewExiftoolMetadataCleanupOperation(
			o.Name,
			o.Params,
			parameterLoaderProvider,
		)
		break
	case "image_convert":
		oh, ohErr = opfactories.NewImageConvertOperation(
			o.Name,
			o.Params,
			parameterLoaderProvider,
		)
		break
	case "s3_upload":
		oh, ohErr = opfactories.NewS3UploadOperation(
			o.Name,
			o.Params,
			parameterLoaderProvider,
		)
		break
	case "filesystem_input_read":
		oh, ohErr = opfactories.NewFilesystemInputReadOperation(
			o.Name,
			o.Params,
			parameterLoaderProvider,
		)
		break
	case "filesystem_input_write":
		oh, ohErr = opfactories.NewFilesystemInputWriteOperation(
			o.Name,
			o.Params,
			parameterLoaderProvider,
		)
		break
	case "filesystem_input_remove":
		oh, ohErr = opfactories.NewFilesystemInputRemoveOperation(o.Name)
		break
	case "command_exec":
		oh, ohErr = opfactories.NewCommandExecOperation(
			o.Name,
			o.Params,
			parameterLoaderProvider,
		)
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

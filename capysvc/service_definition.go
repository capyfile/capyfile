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
	"sync/atomic"
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
) ([]files.ProcessableFile, error) {
	firstOp, lastOp := p.firstAndLastOperations()

	if len(in) > 0 {
		firstOp.operationState.io.in = in
	}

	for i := range p.Operations {
		op := &p.Operations[i]

		handler, handlerErr := op.Handler(ctx)
		if handlerErr != nil {
			return nil, handlerErr
		}

		opHandleOut, opHandlerErr := handler.Handle(op.operationState.io.in, nil, nil)
		if opHandlerErr != nil {
			return opHandleOut, opHandlerErr
		}

		if op.operationState.nextOperation != nil {
			op.operationState.nextOperation.operationState.io.in = opHandleOut
		} else {
			op.operationState.io.out = opHandleOut
		}
	}

	return lastOp.operationState.io.out, nil
}

func (p *Processor) RunOperationsConcurrently(
	ctx Context,
	in []files.ProcessableFile,
	errorCh chan<- operations.OperationError,
	notificationCh chan<- operations.OperationNotification,
) ([]files.ProcessableFile, error) {
	var completeWg sync.WaitGroup

	outputHolder := newOutputHolder()

	for i := range p.Operations {
		completeWg.Add(1)

		op := &p.Operations[i]

		handler, handlerErr := op.Handler(ctx)
		if handlerErr != nil {
			if errorCh != nil {
				errorCh <- operations.NewOperationError(op.Name, handlerErr)
			}

			return outputHolder.Out, handlerErr
		}

		if i == 0 {
			// If this is the first operation, we need to pass the initial input to the operation.
			if len(in) != 0 {
				op.operationState.io.enqueueIn(in...)
			}
		}

		// The below stuff is quite simple. We have 3 goroutines that are running concurrently:
		//   - the first one is passing the operation's output to the next operation's input
		//   - the second one is handling the operation's input
		//   - the third one is checking whether the operation can be completed or not

		// Pass the output to the next operation's input.
		go func(op *Operation) {
			nextOp := op.operationState.nextOperation

			if nextOp == nil {
				// If this is the last operation, then all we need to do is dequeue the
				// output to the final output holder.
				for !op.operationState.Completed {
					op.operationState.io.dequeueOut(
						func(out []files.ProcessableFile) {
							outputHolder.Append(out...)
						},
					)
				}

				return
			}

			for !op.operationState.Completed {
				op.operationState.io.dequeueOut(
					func(out []files.ProcessableFile) {
						nextOp.operationState.io.enqueueIn(out...)
					},
				)
			}
		}(op)

		// Handle the operation's input.
		go func(op *Operation) {
			handleFn := func(in []files.ProcessableFile) (out []files.ProcessableFile) {
				targetIn, skipIn := splitIntoTargetSkip(op.TargetFiles, in)

				opHandleOut, opHandleErr := handler.Handle(targetIn, errorCh, notificationCh)
				if opHandleErr != nil {
					// Perhaps maybe this is something that is related to the specific file,
					// so we can just send an error to the channel and continue.
					if errorCh != nil {
						errorCh <- operations.NewOperationInputError(op.Name, targetIn, opHandleErr)
					}
				}

				if len(skipIn) > 0 {
					for _, pf := range skipIn {
						if notificationCh != nil {
							notificationCh <- operations.NewSkippedOperationNotification(op.Name, op.TargetFiles, &pf)
						}
					}

					out = skipIn
				}

				if len(opHandleOut) > 0 {
					assignCleanupPolicy(op.CleanupPolicy, opHandleOut)

					out = append(out, opHandleOut...)
				}

				return out
			}

			if handler.AllowConcurrency() {
				// If this is a concurrent operation, we run it until it's completed.
				for !op.operationState.Completed {
					op.operationState.io.processIn(handleFn)
				}
			}

			prevOp := op.operationState.prevOperation
			if prevOp != nil {
				// For the operations that are not concurrent, we need to wait for the previous
				// operation to be completed, so we can collect all the input for this operation
				// before run it.
				for !prevOp.operationState.Completed {
					// Wait for the previous operation to be completed.
				}
			}
			op.operationState.io.processIn(handleFn)
		}(op)

		// Complete the operation when no input is expected for it.
		go func(op *Operation) {
			defer completeWg.Done()

			// The operation must be run at least once. Only after this it makes sense to
			// check whether the operation can be completed or not.
			for op.operationState.io.procCnt.Load() == 0 {
			}

			prevOp := op.operationState.prevOperation
			for !op.operationState.Completed {
				// If this is not the first operation, we need to ensure that the previous one is completed.
				if prevOp == nil || prevOp.operationState.Completed {
					if op.operationState.io.isEmpty() {
						op.operationState.complete()
						return
					}
				}
			}
		}(op)
	}

	completeWg.Wait()

	return outputHolder.Out, nil
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

		op.Reset()

		if prevOp != nil {
			prevOp.operationState.nextOperation = op
			prevOp.operationState.nextOperation.operationState.prevOperation = prevOp
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

	operationState *operationState
}

type operationState struct {
	// Manages the input and output of the operation.
	io *ioManager

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

// Provides the basic methods to manage the operation's input and output.
type ioManager struct {
	inOutLock *sync.RWMutex
	in        []files.ProcessableFile
	out       []files.ProcessableFile

	procCnt *atomic.Uint32
}

func (m *ioManager) isEmpty() bool {
	m.inOutLock.RLock()
	defer m.inOutLock.RUnlock()

	return len(m.in) == 0 && len(m.out) == 0
}

func (m *ioManager) enqueueIn(pf ...files.ProcessableFile) {
	m.inOutLock.Lock()
	defer m.inOutLock.Unlock()

	m.in = append(m.in, pf...)
}

func (m *ioManager) processIn(
	f func(in []files.ProcessableFile) (out []files.ProcessableFile),
) {
	m.inOutLock.Lock()
	defer m.inOutLock.Unlock()

	out := f(m.in)

	m.in = nil
	m.out = append(m.out, out...)

	m.procCnt.Add(1)
}

func (m *ioManager) dequeueOut(
	f func(out []files.ProcessableFile),
) {
	m.inOutLock.Lock()
	defer m.inOutLock.Unlock()

	if len(m.out) == 0 {
		return
	}

	f(m.out)

	m.out = nil
}

func (o *Operation) Reset() {
	o.operationState = &operationState{
		io: &ioManager{
			inOutLock: &sync.RWMutex{},
			procCnt:   &atomic.Uint32{},
		},
		completedLock: &sync.Mutex{},
		handlerLock:   &sync.Mutex{},
	}
}

func (o *operationState) complete() {
	o.completedLock.Lock()
	defer o.completedLock.Unlock()

	o.Completed = true
}

func (o *Operation) Handler(ctx Context) (operations.OperationHandler, error) {
	o.operationState.handlerLock.Lock()
	defer o.operationState.handlerLock.Unlock()

	// If the handler is already created, return it.
	if o.operationState.handler != nil {
		return o.operationState.handler, nil
	}

	var oh operations.OperationHandler
	var ohErr error

	parameterLoaderProvider, parameterLoaderProviderErr := ctx.ParameterLoaderProvider()
	if parameterLoaderProviderErr != nil {
		return nil, parameterLoaderProviderErr
	}

	switch o.Name {
	case "http_multipart_form_input_read":
		req := ctx.Request()
		if req == nil {
			return nil, errors.New("http request is not available in the given context")
		}

		oh, ohErr = opfactories.NewHttpMultipartFormInputReadOperation(o.Name, ctx.Request())
		break
	case "http_octet_stream_input_read":
		req := ctx.Request()
		if req == nil {
			return nil, errors.New("http request is not available in the given context")
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
		return nil, fmt.Errorf("unknown operation \"%s\"", o.Name)
	}

	if ohErr != nil {
		return nil, ohErr
	}

	return oh, nil
}

func splitIntoTargetSkip(targetPolicy string, in []files.ProcessableFile) (target, skip []files.ProcessableFile) {
	if targetPolicy == OperationTargetFilesAll {
		return in, []files.ProcessableFile{}
	}

	if targetPolicy == OperationTargetFilesWithoutErrors {
		for _, pf := range in {
			if !pf.HasFileProcessingError() {
				target = append(target, pf)
			} else {
				skip = append(skip, pf)
			}
		}
	} else if targetPolicy == OperationTargetFilesWithErrors {
		for _, pf := range in {
			if pf.HasFileProcessingError() {
				target = append(target, pf)
			} else {
				skip = append(skip, pf)
			}
		}
	}

	return target, skip
}

func assignCleanupPolicy(cleanupPolicy string, in []files.ProcessableFile) {
	for i := range in {
		pf := &in[i]

		switch cleanupPolicy {
		case OperationCleanupPolicyKeepFiles:
			pf.KeepOnFreeResources()
			break
		case OperationCleanupPolicyRemoveFiles:
			pf.RemoveOnFreeResources()
			break
		}
	}
}

type outHolder struct {
	outLock *sync.Mutex
	Out     []files.ProcessableFile
}

func newOutputHolder() *outHolder {
	return &outHolder{
		outLock: &sync.Mutex{},
	}
}

func (o *outHolder) Append(pf ...files.ProcessableFile) {
	o.outLock.Lock()
	defer o.outLock.Unlock()

	o.Out = append(o.Out, pf...)
}

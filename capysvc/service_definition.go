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
	"time"
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
	errorCh chan<- operations.OperationError,
	notificationCh chan<- operations.OperationNotification,
) (out []files.ProcessableFile, err error) {
	proc := s.FindProcessor(processorName)
	if proc == nil {
		return out, capyerr.NewProcessorNotFoundError(processorName)
	}

	initOpsErr := proc.InitOperations(ctx)
	if initOpsErr != nil {
		return out, initOpsErr
	}

	return proc.RunOperations(ctx, in, errorCh, notificationCh)
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
	errorCh chan<- operations.OperationError,
	notificationCh chan<- operations.OperationNotification,
) ([]files.ProcessableFile, error) {
	firstOp, lastOp := p.firstAndLastOperations()

	if len(in) > 0 {
		firstOp.operationState.io.in = in
	}

	for i := range p.Operations {
		op := &p.Operations[i]

		handler, handlerErr := op.Handler(ctx)
		if handlerErr != nil {
			if errorCh != nil {
				errorCh <- operations.NewOperationError(op.Name, handlerErr)
			}

			return nil, handlerErr
		}

		targetIn, skipIn := splitIntoTargetSkip(op.TargetFiles, op.operationState.io.in)

		opHandleOut, opHandlerErr := handler.Handle(targetIn, errorCh, notificationCh)
		if opHandlerErr != nil {
			if errorCh != nil {
				errorCh <- operations.NewOperationInputError(op.Name, op.operationState.io.in, opHandlerErr)
			}

			return opHandleOut, opHandlerErr
		}

		assignCleanupPolicy(op.CleanupPolicy, opHandleOut)

		if op.operationState.nextOperation != nil {
			op.operationState.nextOperation.operationState.io.in = append(opHandleOut, skipIn...)
		} else {
			op.operationState.io.out = append(opHandleOut, skipIn...)
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
	var (
		inOutWg    sync.WaitGroup
		procWg     sync.WaitGroup
		completeWg sync.WaitGroup
	)

	firstOp, lastOp := p.firstAndLastOperations()

	if len(in) > 0 {
		firstOp.operationState.io.in = in
	}

	for i := range p.Operations {
		op := &p.Operations[i]

		handler, handlerErr := op.Handler(ctx)
		if handlerErr != nil {
			if errorCh != nil {
				errorCh <- operations.NewOperationError(op.Name, handlerErr)
			}

			return nil, handlerErr
		}

		// The below stuff is quite simple. We have 3 goroutines that are running concurrently:
		//   - the first one is passing the operation's output to the next operation's input
		//   - the second one is handling the operation's input
		//   - the third one is checking whether the operation can be completed or not

		// Pass the output to the next operation's input.
		inOutWg.Add(1)
		go func(op *Operation) {
			defer inOutWg.Done()

			nextOp := op.operationState.nextOperation

			if nextOp == nil {
				// If this is the last operation, we don't need to dequeue the output anywhere.
				return
			}

			for !op.operationState.Completed {
				if op.operationState.io.isOutputQueueEmpty() {
					op.IOTick()

					continue
				}

				op.operationState.io.dequeueOutput(
					func(out []files.ProcessableFile) {
						nextOp.operationState.io.enqueueInput(out...)
					},
				)

				op.IOTick()
			}
		}(op)

		// Handle the operation's input.
		procWg.Add(1)
		go func(op *Operation) {
			defer procWg.Done()

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
					if op.operationState.io.procCnt > 0 && op.operationState.io.isInputQueueEmpty() {
						op.HandlerTick()

						continue
					}

					op.operationState.io.process(op.MaxPacketSize, handleFn)

					op.HandlerTick()
				}

				return
			}

			prevOp := op.operationState.prevOperation
			if prevOp != nil {
				// For the operations that are not concurrent, we need to wait for the previous
				// operation to be completed, so we can collect all the input for this operation
				// before run it.
				for !prevOp.operationState.Completed {
					// Wait for the previous operation to be completed.
					op.HandlerTick()
				}
			}
			op.operationState.io.process(0, handleFn)
		}(op)

		// Complete the operation when no input is expected for it.
		completeWg.Add(1)
		go func(op *Operation) {
			defer completeWg.Done()

			// The operation must be run at least once. Only after this it makes sense to
			// check whether the operation can be completed or not.
			for op.operationState.io.procCnt == 0 {
				op.StatusTick()
			}

			prevOp := op.operationState.prevOperation
			if prevOp != nil {
				// If this is not the first operation, we need to ensure that the previous one is completed.
				for !prevOp.operationState.Completed {
					op.StatusTick()
				}
			}

			nextOp := op.operationState.nextOperation
			if nextOp != nil {
				for !op.operationState.Completed {
					if op.operationState.io.isQueueEmpty() {
						op.operationState.complete()
						return
					}

					op.StatusTick()
				}
			}

			for !op.operationState.Completed {
				if op.operationState.io.isInputQueueEmpty() {
					op.operationState.complete()
					return
				}

				op.StatusTick()
			}
		}(op)
	}

	inOutWg.Wait()
	procWg.Wait()
	completeWg.Wait()

	return lastOp.operationState.io.out, nil
}

func (s *Service) RunProcessorConcurrentlyWithEventBasedConcurrencyAlgorithm(
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

	return proc.RunOperationsConcurrentlyWithEventBasedConcurrencyAlgorithm(
		ctx,
		in,
		errorCh,
		notificationCh,
	)
}

func (p *Processor) RunOperationsConcurrentlyWithEventBasedConcurrencyAlgorithm(
	ctx Context,
	in []files.ProcessableFile,
	errorCh chan<- operations.OperationError,
	notificationCh chan<- operations.OperationNotification,
) ([]files.ProcessableFile, error) {
	type opCh struct {
		inputCh    chan struct{}
		processCh  chan struct{}
		outputCh   chan struct{}
		completeCh chan struct{}
	}
	opChs := make([]opCh, len(p.Operations))

	completeWg := &sync.WaitGroup{}

	for i := range p.Operations {
		op := &p.Operations[i]

		handler, handlerErr := op.Handler(ctx)
		if handlerErr != nil {
			if errorCh != nil {
				errorCh <- operations.NewOperationError(op.Name, handlerErr)
			}

			return nil, handlerErr
		}

		opChs[i] = opCh{
			inputCh:    make(chan struct{}),
			processCh:  make(chan struct{}),
			outputCh:   make(chan struct{}),
			completeCh: make(chan struct{}),
		}

		go func(opIdx int) {
			for {
				select {
				case <-opChs[opIdx].inputCh:
					// If this is not a concurrent operation, we just stack the input,
					// so it will be processed when the previous operation is completed.
					if !handler.AllowConcurrency() {
						// If this is not the first operation, and the previous
						// operation is not completed, then we can just return, so
						// the current operation will continue to collect the input.
						prevOp := op.operationState.prevOperation
						if prevOp != nil && !prevOp.operationState.Completed {
							continue
						}
					}

					opChs[opIdx].processCh <- struct{}{}
				}
			}
		}(i)

		go func(opIdx int) {
			for {
				select {
				case <-opChs[opIdx].processCh:
					op.operationState.io.process(0, func(in []files.ProcessableFile) (out []files.ProcessableFile) {
						targetIn, skipIn := splitIntoTargetSkip(op.TargetFiles, in)

						var handleOut []files.ProcessableFile
						var handleErr error

						if op.MaxPacketSize > 0 {
							// Here we are chunking the input into the pieces based on the max packet size.
							var chunks [][]files.ProcessableFile
							for op.MaxPacketSize < len(targetIn) {
								targetIn, chunks = targetIn[op.MaxPacketSize:], append(chunks, targetIn[0:op.MaxPacketSize:op.MaxPacketSize])
							}
							chunks = append(chunks, targetIn)

							// Right now we process the chunks in parallel.
							wg := &sync.WaitGroup{}
							l := &sync.Mutex{}
							for _, chunk := range chunks {
								wg.Add(1)
								go func(chunk []files.ProcessableFile) {
									defer wg.Done()

									chunkHandleOut, chunkHandleErr := handler.Handle(chunk, errorCh, notificationCh)
									if chunkHandleErr != nil {
										if errorCh != nil {
											errorCh <- operations.NewOperationInputError(op.Name, chunk, chunkHandleErr)
										}

										return
									}

									l.Lock()
									handleOut = append(handleOut, chunkHandleOut...)
									l.Unlock()
								}(chunk)
							}
							wg.Wait()
						} else {
							handleOut, handleErr = handler.Handle(targetIn, errorCh, notificationCh)
							if handleErr != nil {
								// Perhaps maybe this is something that is related to the specific file,
								// so we can just send an error to the channel and continue.
								if errorCh != nil {
									errorCh <- operations.NewOperationInputError(op.Name, targetIn, handleErr)
								}

								return out
							}
						}

						if len(skipIn) > 0 {
							if notificationCh != nil {
								for _, pf := range skipIn {
									notificationCh <- operations.NewSkippedOperationNotification(op.Name, op.TargetFiles, &pf)
								}
							}

							out = skipIn
						}

						if len(handleOut) > 0 {
							assignCleanupPolicy(op.CleanupPolicy, handleOut)

							out = append(out, handleOut...)
						}

						return out
					})

					opChs[opIdx].outputCh <- struct{}{}
				}
			}
		}(i)

		go func(opIdx int) {
			for {
				select {
				case <-opChs[opIdx].outputCh:
					// Pass the output to the next operation's input.
					nextOp := op.operationState.nextOperation
					if nextOp != nil {
						op.operationState.io.dequeueOutput(
							func(out []files.ProcessableFile) {
								nextOp.operationState.io.enqueueInput(out...)
							},
						)

						opChs[opIdx].completeCh <- struct{}{}
						opChs[opIdx+1].inputCh <- struct{}{}
					} else {
						opChs[opIdx].completeCh <- struct{}{}
					}
				}
			}
		}(i)

		completeWg.Add(1)
		go func(opIdx int) {
			for {
				select {
				case <-opChs[opIdx].completeCh:
					// If this is not the first operation, and the previous
					// operation is not completed, then we can just return, so
					// the current operation will continue to collect the input.
					prevOp := op.operationState.prevOperation
					if prevOp != nil && !prevOp.operationState.Completed {
						continue
					}

					// If this is not the last operation, we need to wait while all
					// the input/output is processed.
					nextOp := op.operationState.nextOperation
					if nextOp != nil {
						if op.operationState.io.isQueueEmpty() {
							op.operationState.complete()
							completeWg.Done()
						}

						continue
					}

					// If this is the last operation, we need to wait while the input
					// is processed.
					if op.operationState.io.isInputQueueEmpty() {
						op.operationState.complete()
						completeWg.Done()
					}
				}
			}
		}(i)
	}

	firstOp, lastOp := p.firstAndLastOperations()

	firstOp.operationState.io.enqueueInput(in...)
	opChs[0].inputCh <- struct{}{}

	completeWg.Wait()

	return lastOp.operationState.io.out, nil
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

		if op.IOTickDelay != "" {
			d, dErr := time.ParseDuration(op.IOTickDelay)
			if dErr != nil {
				return dErr
			}

			op.ioTickDelayDuration = d
		}
		if op.HandlerTickDelay != "" {
			d, dErr := time.ParseDuration(op.HandlerTickDelay)
			if dErr != nil {
				return dErr
			}

			op.handlerTickDelayDuration = d
		}
		if op.StatusTickDelay != "" {
			d, dErr := time.ParseDuration(op.StatusTickDelay)
			if dErr != nil {
				return dErr
			}

			op.statusTickDelayDuration = d
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

	// MaxPacketSize is the parameter that defines the maximum size of the
	// packet that is passed to the operation handler. Packet in this
	// context is the input (files) that is passed to the operation handler
	// at once.
	// For example, 10 means the operation handler will be processing not
	// more than 10 files at once.
	MaxPacketSize int `json:"maxPacketSize" yaml:"maxPacketSize"`

	// By default, capyfile provides maximum performance. But all this performance comes
	// with a cost of CPU usage. So if you want to reduce the CPU usage, you can set the
	// tick delays for the operations. In general, it makes sense to set the tick delays
	// if you pipeline works with the large files.

	// IOTickDelay is the parameter that defines the tick delay of the IO
	// operations, such as moving the input between the operations. It can be
	// any duration string that is supported by the time.ParseDuration function,
	// for example, "1s", "1ms", "10μs", "1.5h", etc.
	IOTickDelay         string `json:"ioTickDelay" yaml:"ioTickDelay"`
	ioTickDelayDuration time.Duration
	// HandlerTickDelay is the parameter that defines the tick delay of the
	// input handler in the concurrent mode. It can be any duration string that
	// is supported by the time.ParseDuration function, for example, "1s",
	// "1ms", "10μs", "1.5h", etc.
	HandlerTickDelay         string `json:"handlerTickDelay" yaml:"handlerTickDelay"`
	handlerTickDelayDuration time.Duration
	// StatusTickDelay is the parameter that defines the tick delay of the
	// operation status update. It can be any duration string that
	// is supported by the time.ParseDuration function, for example, "1s",
	// "1ms", "10μs", "1.5h", etc.
	StatusTickDelay         string `json:"statusTickDelay" yaml:"statusTickDelay"`
	statusTickDelayDuration time.Duration

	operationState *operationState
}

func (o *Operation) IOTick() {
	if o.ioTickDelayDuration == 0 {
		return
	}

	time.Sleep(o.ioTickDelayDuration)
}

func (o *Operation) HandlerTick() {
	if o.handlerTickDelayDuration == 0 {
		return
	}

	time.Sleep(o.handlerTickDelayDuration)
}

func (o *Operation) StatusTick() {
	if o.statusTickDelayDuration == 0 {
		return
	}

	time.Sleep(o.statusTickDelayDuration)
}

func (o *Operation) Reset() {
	o.operationState = &operationState{
		io: &ioManager{
			inOutLock: &sync.RWMutex{},
			procCnt:   0,
		},
		completedLock: &sync.Mutex{},
		handlerLock:   &sync.Mutex{},
	}
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
		oh, ohErr = opfactories.NewFilesystemInputRemoveOperation(
			o.Name,
			o.Params,
			parameterLoaderProvider,
		)
		break
	case "input_forget":
		oh, ohErr = opfactories.NewInputForgetOperation(o.Name)
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

	o.operationState.handler = oh

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

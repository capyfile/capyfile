package operations

import (
	"capyfile/files"
	"sync"
)

type OperationHandler interface {
	OperationName() string
	// AllowConcurrency returns true if the operation can be executed in a concurrent manner.
	//
	// You should not allow this if, for example:
	//  - the operation required all output from the previous operation
	//  - the operation accepts empty input and produces non-empty output
	//  - the operation should block the execution of the next operation until it
	//    finishes
	AllowConcurrency() bool
	// Handle handles the processable files.
	//
	// The in is the slice of processable files to be processed. The reason why
	// the slice is used here instead of the individual processable file is because,
	// is gives a lot more options of the operations to implement. For example,
	// the operations can:
	//   - process input in batches
	//   - process input in parallel
	//   - concatenate the input
	//   - have an empty input which can be useful for the operations that retrieve
	//     the input from some sources (filesystem, network, storage, etc.)
	//
	// The errorCh channel is used to send the internal errors encountered during
	// specific processable file handling. This is the type of error that the
	// host/server may have to pay attention to. This type of error may not be safe
	// to be returned to the client.
	//
	// notificationCh channel is used to send the notifications related to the
	// processable file handling. For example, it can be the notification like
	// "processing started", "processing finished", "processing failed", etc.
	//
	// The out is the slice of processable files that were processed (successfully or not).
	//
	// The err is the error that is related to the operation itself. For example,
	// the operation is not configured properly or some dependency is missing. This
	// shouldn't be the error related to the individual processable file.
	Handle(
		in []files.ProcessableFile,
		errorCh chan<- OperationError,
		notificationCh chan<- OperationNotification,
	) (out []files.ProcessableFile, err error)
}

func newOutputHolder() *outputHolder {
	return &outputHolder{
		outLock: sync.Mutex{},
	}
}

type outputHolder struct {
	Out     []files.ProcessableFile
	outLock sync.Mutex
}

func (oh *outputHolder) AppendToOut(pf *files.ProcessableFile) {
	oh.outLock.Lock()
	oh.Out = append(oh.Out, *pf)
	oh.outLock.Unlock()
}

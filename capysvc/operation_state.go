package capysvc

import (
	"capyfile/operations"
	"sync"
)

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

func (o *operationState) complete() {
	o.completedLock.Lock()
	defer o.completedLock.Unlock()

	o.Completed = true
}

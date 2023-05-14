package capyerr

import "fmt"

type Error interface {
	error

	Code() string
	Message() string
	OriginalError() error
}

func New(code, message string, origErr error) Error {
	return &baseError{
		code:    code,
		message: message,
		errs:    []error{origErr},
	}
}

type baseError struct {
	code    string
	message string
	// Allows us building chained errors.
	errs []error
}

func (e *baseError) Error() string {
	return e.message
}

func (e *baseError) Code() string {
	return e.code
}

func (e *baseError) Message() string {
	return e.message
}

func (e *baseError) OriginalError() error {
	return e.errs[0]
}

type OperationConfigurationType struct {
	code    string
	message string
	// Allows us building chained errors.
	errs []error
}

func NewOperationConfigurationError(code, message string, origErr error) *OperationConfigurationType {
	return &OperationConfigurationType{
		code:    code,
		message: message,
		errs:    []error{origErr},
	}
}

func (e *OperationConfigurationType) Error() string {
	return e.message
}

func (e *OperationConfigurationType) Code() string {
	return e.code
}

func (e *OperationConfigurationType) Message() string {
	return e.message
}

func (e *OperationConfigurationType) OriginalError() error {
	return e.errs[0]
}

type ProcessorNotFoundType struct {
	code    string
	message string
	// Allows us building chained errors.
	errs []error
}

func NewProcessorNotFoundError(processorName string) *ProcessorNotFoundType {
	return &ProcessorNotFoundType{
		code:    "PROCESSOR_NOT_FOUND",
		message: fmt.Sprintf("processor %s not found", processorName),
	}
}

func (e *ProcessorNotFoundType) Error() string {
	return e.message
}

func (e *ProcessorNotFoundType) Code() string {
	return e.code
}

func (e *ProcessorNotFoundType) Message() string {
	return e.message
}

func (e *ProcessorNotFoundType) OriginalError() error {
	return nil
}

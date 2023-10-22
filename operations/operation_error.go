package operations

import "capyfile/files"

type OperationError struct {
	OperationName   string
	In              []files.ProcessableFile
	ProcessableFile *files.ProcessableFile
	Err             error
}

type OperationErrorBuilder struct {
	OperationName string
}

func (eb *OperationErrorBuilder) InputError(in []files.ProcessableFile, err error) OperationError {
	return OperationError{
		OperationName: eb.OperationName,
		In:            in,
		Err:           err,
	}
}

func (eb *OperationErrorBuilder) ProcessableFileError(pf *files.ProcessableFile, err error) OperationError {
	return OperationError{
		OperationName:   eb.OperationName,
		ProcessableFile: pf,
		Err:             err,
	}
}

func (eb *OperationErrorBuilder) Error(err error) OperationError {
	return OperationError{
		OperationName: eb.OperationName,
		Err:           err,
	}
}

func NewOperationInputError(operationName string, in []files.ProcessableFile, err error) OperationError {
	return OperationError{
		OperationName: operationName,
		In:            in,
		Err:           err,
	}
}

func NewOperationError(operationName string, err error) OperationError {
	return OperationError{
		OperationName: operationName,
		Err:           err,
	}
}

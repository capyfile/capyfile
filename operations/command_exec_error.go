package operations

import "capyfile/files"

const ErrorCodeCommandTemplateCanNotBeRendered = "COMMAND_TEMPLATE_CAN_NOT_BE_RENDERED"

func NewCommandTemplateCanNotBeRenderedError(origErr error) *CommandTemplateCanNotBeRenderedError {
	return &CommandTemplateCanNotBeRenderedError{
		Data: &CommandTemplateCanNotBeRenderedErrorData{
			OrigErr: origErr,
		},
	}
}

type CommandTemplateCanNotBeRenderedError struct {
	files.FileProcessingError
	Data *CommandTemplateCanNotBeRenderedErrorData
}

type CommandTemplateCanNotBeRenderedErrorData struct {
	OrigErr error
}

func (e *CommandTemplateCanNotBeRenderedError) Code() string {
	return ErrorCodeCommandTemplateCanNotBeRendered
}

func (e *CommandTemplateCanNotBeRenderedError) Error() string {
	return "command template can not be rendered"
}

const ErrorCodeCommandExecutionError = "COMMAND_EXECUTION_ERROR"

func NewCommandExecutionError(origErr error) *CommandExecutionError {
	return &CommandExecutionError{
		Data: &CommandExecutionErrorData{
			OrigErr: origErr,
		},
	}
}

type CommandExecutionError struct {
	files.FileProcessingError
	Data *CommandExecutionErrorData
}

type CommandExecutionErrorData struct {
	OrigErr error
}

func (e *CommandExecutionError) Code() string {
	return ErrorCodeCommandExecutionError
}

func (e *CommandExecutionError) Error() string {
	return "command execution error"
}

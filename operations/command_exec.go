package operations

import (
	"bytes"
	"capyfile/capyfs"
	"capyfile/files"
	"errors"
	"html/template"
	"os/exec"
	"sync"
)

type CommandExecOperation struct {
	Name            string
	Params          *CommandExecOperationParams
	CommandExecutor CommandExecutor
}

func (o *CommandExecOperation) OperationName() string {
	return o.Name
}

func (o *CommandExecOperation) AllowConcurrency() bool {
	// We assume that if the command does not support parallel execution, it also
	// does not support concurrency.
	return o.Params.AllowParallelExecution
}

type CommandExecOperationParams struct {
	CommandName            string
	CommandArgs            []string
	OutputFileDestination  string
	AllowParallelExecution bool
}

type templateData struct {
	// Absolute path to the file being processed. For example: /home/user/file.bin
	AbsolutePath string
	// Filename with extension. For example: file.bin
	Filename string
	// Filename without extension. For example: file
	Basename string
	// File extension. For example: bin
	Extension string

	// The same stuff but for the original file.
	OriginalAbsolutePath string
	OriginalFilename     string
	OriginalBasename     string
	OriginalExtension    string
}

func (o *CommandExecOperation) Handle(
	in []files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (out []files.ProcessableFile, err error) {
	commandExecutorInitErr := o.initCommandExecutor()
	if commandExecutorInitErr != nil {
		if errorCh != nil {
			errorCh <- o.errorBuilder().Error(commandExecutorInitErr)
		}

		return nil, commandExecutorInitErr
	}

	outHolder := newOutputHolder()

	var wg sync.WaitGroup

	execFunc := func(pf *files.ProcessableFile) {
		defer wg.Done()

		if notificationCh != nil {
			notificationCh <- o.notificationBuilder().Started("command execution has started", pf)
		}

		// First we want to render all the templates (command name and args) that we need
		// to execute the command.

		tmplData := templateData{
			AbsolutePath: pf.FileAbsolutePath(),
			Filename:     pf.Filename(),
			Basename:     pf.FileBasename(),
			Extension:    pf.FileExtension(),
		}
		if pf.OriginalProcessableFile != nil {
			tmplData.OriginalAbsolutePath = pf.OriginalProcessableFile.FileAbsolutePath()
			tmplData.OriginalFilename = pf.OriginalProcessableFile.Filename()
			tmplData.OriginalBasename = pf.OriginalProcessableFile.FileBasename()
			tmplData.OriginalExtension = pf.OriginalProcessableFile.FileExtension()
		}

		var commandNameBuf bytes.Buffer
		commandNameTemplate, tmplParseErr := template.New("commandName").Parse(o.Params.CommandName)
		if tmplParseErr != nil {
			pf.SetFileProcessingError(
				NewCommandTemplateCanNotBeRenderedError(tmplParseErr),
			)

			if errorCh != nil {
				errorCh <- o.errorBuilder().Error(tmplParseErr)
			}
			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Failed(
					"command name template can not be parsed", pf, tmplParseErr)
			}

			outHolder.AppendToOut(pf)

			return
		}
		tmplExecErr := commandNameTemplate.Execute(&commandNameBuf, tmplData)
		if tmplExecErr != nil {
			pf.SetFileProcessingError(
				NewCommandTemplateCanNotBeRenderedError(tmplExecErr),
			)

			if errorCh != nil {
				errorCh <- o.errorBuilder().Error(tmplExecErr)
			}
			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Failed(
					"command name template can not be rendered", pf, tmplExecErr)
			}

			outHolder.AppendToOut(pf)

			return
		}

		var commandArgs []string
		for _, arg := range o.Params.CommandArgs {
			var commandArgBuf bytes.Buffer
			commandArgTemplate, argTmplParseErr := template.New("commandArg").Parse(arg)
			if argTmplParseErr != nil {
				pf.SetFileProcessingError(
					NewCommandTemplateCanNotBeRenderedError(argTmplParseErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().Error(argTmplParseErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"command argument template can not be parsed", pf, argTmplParseErr)
				}

				outHolder.AppendToOut(pf)

				return
			}
			argTmplExecErr := commandArgTemplate.Execute(&commandArgBuf, tmplData)
			if argTmplExecErr != nil {
				pf.SetFileProcessingError(
					NewCommandTemplateCanNotBeRenderedError(argTmplExecErr),
				)

				if errorCh != nil {
					errorCh <- o.errorBuilder().Error(argTmplExecErr)
				}
				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"command argument template can not be rendered", pf, argTmplExecErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			commandArgs = append(commandArgs, commandArgBuf.String())
		}

		// Now when all the templates are rendered, we can execute the command.

		output, execErr := o.CommandExecutor.Execute(commandNameBuf.String(), commandArgs...)
		if execErr != nil {
			pf.SetFileProcessingError(
				NewCommandExecutionError(execErr),
			)

			if errorCh != nil {
				errorCh <- o.errorBuilder().Error(execErr)
				errorCh <- o.errorBuilder().Error(
					errors.New(string(output)))
			}
			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Failed(
					"command execution has failed", pf, execErr)
			}

			outHolder.AppendToOut(pf)

			return
		}

		// If the command has finished successfully, we need to render the output file destination
		// template and create a processable file from it.

		outputFileTmpl, outputFileTmplErr := template.New("outputFile").Parse(o.Params.OutputFileDestination)
		if outputFileTmplErr != nil {
			pf.SetFileProcessingError(
				NewCommandTemplateCanNotBeRenderedError(outputFileTmplErr),
			)

			if errorCh != nil {
				errorCh <- o.errorBuilder().Error(outputFileTmplErr)
			}
			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Failed(
					"output file destination template can not be parsed", pf, outputFileTmplErr)
			}

			outHolder.AppendToOut(pf)

			return
		}
		var outputFileBuf bytes.Buffer
		outputFileTmplExecErr := outputFileTmpl.Execute(&outputFileBuf, tmplData)
		if outputFileTmplExecErr != nil {
			pf.SetFileProcessingError(
				NewCommandTemplateCanNotBeRenderedError(outputFileTmplExecErr),
			)

			if errorCh != nil {
				errorCh <- o.errorBuilder().Error(outputFileTmplExecErr)
			}
			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Failed(
					"output file destination template can not be rendered", pf, outputFileTmplExecErr)
			}

			outHolder.AppendToOut(pf)

			return
		}

		file, fileOpenErr := capyfs.Filesystem.Open(outputFileBuf.String())
		if fileOpenErr != nil {
			pf.SetFileProcessingError(
				NewFileIsUnreadableError(fileOpenErr),
			)

			if errorCh != nil {
				errorCh <- o.errorBuilder().Error(fileOpenErr)
			}
			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Failed(
					"can not open the command output file", pf, fileOpenErr)
			}

			outHolder.AppendToOut(pf)

			return
		}

		pf.ReplaceFile(file)

		if notificationCh != nil {
			notificationCh <- o.notificationBuilder().Finished("command execution has finished", pf)
		}

		outHolder.AppendToOut(pf)
	}

	for i := range in {
		wg.Add(1)

		pf := &in[i]
		if o.Params.AllowParallelExecution {
			go execFunc(pf)
		} else {
			execFunc(pf)
		}
	}

	wg.Wait()

	return outHolder.Out, nil
}

func (o *CommandExecOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.OperationName(),
	}
}

func (o *CommandExecOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.OperationName(),
	}
}

func (o *CommandExecOperation) initCommandExecutor() error {
	if o.CommandExecutor == nil {
		o.CommandExecutor = CommandExecutorFunc(
			func(name string, arg ...string) (output []byte, err error) {
				cmd := exec.Command(name, arg...)

				return cmd.CombinedOutput()
			},
		)
	}

	return nil
}

type CommandExecutor interface {
	Execute(name string, arg ...string) (output []byte, err error)
}

type CommandExecutorFunc func(name string, arg ...string) (output []byte, err error)

func (f CommandExecutorFunc) Execute(name string, arg ...string) (output []byte, err error) {
	return f(name, arg...)
}

package operations

import (
	"bytes"
	"capyfile/capyfs"
	"capyfile/files"
	"errors"
	"fmt"
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
	// File extension. For example: .bin
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

		var tmplData templateData

		if pf != nil {
			absolutePath, absolutePathErr := pf.FileAbsolutePath()
			if absolutePathErr != nil {
				if errorCh != nil {
					errorCh <- o.errorBuilder().Error(absolutePathErr)
				}

				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"can not get the absolute path to the file", pf, absolutePathErr)
				}

				outHolder.AppendToOut(pf)

				return
			}

			tmplData = templateData{
				AbsolutePath:         absolutePath,
				Filename:             pf.Filename(),
				Basename:             pf.FileBasename(),
				Extension:            pf.FileExtension(),
				OriginalAbsolutePath: absolutePath,
				OriginalFilename:     pf.Filename(),
				OriginalBasename:     pf.FileBasename(),
				OriginalExtension:    pf.FileExtension(),
			}
			if pf.OriginalProcessableFile != nil {
				originalAbsolutePath, originalAbsolutePathErr := pf.OriginalProcessableFile.FileAbsolutePath()
				if originalAbsolutePathErr != nil {
					if errorCh != nil {
						errorCh <- o.errorBuilder().Error(originalAbsolutePathErr)
					}

					if notificationCh != nil {
						notificationCh <- o.notificationBuilder().Failed(
							"can not get the absolute path to the original file", pf, originalAbsolutePathErr)
					}

					outHolder.AppendToOut(pf)

					return
				}

				tmplData.OriginalAbsolutePath = originalAbsolutePath
				tmplData.OriginalFilename = pf.OriginalProcessableFile.Filename()
				tmplData.OriginalBasename = pf.OriginalProcessableFile.FileBasename()
				tmplData.OriginalExtension = pf.OriginalProcessableFile.FileExtension()
			}
		}

		cmdName, cmdNameErr := o.renderTemplate(
			"command name",
			o.Params.CommandName,
			tmplData,
			pf,
			errorCh,
			notificationCh,
		)
		if cmdNameErr != nil {
			if pf != nil {
				outHolder.AppendToOut(pf)
			}

			return
		}

		var cmdArgs []string
		for _, arg := range o.Params.CommandArgs {
			cmdArg, cmdArgErr := o.renderTemplate(
				"command argument",
				arg,
				tmplData,
				pf,
				errorCh,
				notificationCh,
			)
			if cmdArgErr != nil {
				if pf != nil {
					outHolder.AppendToOut(pf)
				}

				return
			}

			cmdArgs = append(cmdArgs, cmdArg)
		}

		// Now when all the templates are rendered, we can execute the command.

		output, execErr := o.CommandExecutor.Execute(cmdName, cmdArgs...)
		if execErr != nil {
			if errorCh != nil {
				errorCh <- o.errorBuilder().Error(execErr)
				errorCh <- o.errorBuilder().Error(
					errors.New(string(output)))
			}

			if pf != nil {
				pf.SetFileProcessingError(
					NewCommandExecutionError(execErr),
				)

				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"command execution has failed", pf, execErr)
				}

				outHolder.AppendToOut(pf)
			}

			return
		}

		// We should consider the case when the command does not produce any output file.
		// In such case we just add the input file to the output if any.
		if o.Params.OutputFileDestination == "" {
			if pf != nil {
				outHolder.AppendToOut(pf)
			}

			return
		}

		// If we expect some output from the command, we need to render the output file
		// destination template and create a processable file from it.

		outputFile, outputFileErr := o.renderTemplate(
			"output file destination",
			o.Params.OutputFileDestination,
			tmplData,
			pf,
			errorCh,
			notificationCh,
		)
		if outputFileErr != nil {
			if pf != nil {
				outHolder.AppendToOut(pf)
			}

			return
		}

		file, fileOpenErr := capyfs.Filesystem.Open(outputFile)
		if fileOpenErr != nil {
			if errorCh != nil {
				errorCh <- o.errorBuilder().Error(fileOpenErr)
			}

			if pf != nil {
				pf.SetFileProcessingError(
					NewFileIsUnreadableError(fileOpenErr),
				)

				if notificationCh != nil {
					notificationCh <- o.notificationBuilder().Failed(
						"can not open the command output file", pf, fileOpenErr)
				}

				outHolder.AppendToOut(pf)
			}

			return
		}

		if pf != nil {
			pf.ReplaceFile(file.Name())
		} else {
			newPf := files.NewProcessableFile(file.Name())
			pf = &newPf
		}

		if notificationCh != nil {
			notificationCh <- o.notificationBuilder().Finished("command execution has finished", pf)
		}

		outHolder.AppendToOut(pf)
	}

	if len(in) == 0 {
		wg.Add(1)
		execFunc(nil)
	} else {
		for i := range in {
			wg.Add(1)

			pf := &in[i]
			if o.Params.AllowParallelExecution {
				go execFunc(pf)
			} else {
				execFunc(pf)
			}
		}
	}

	wg.Wait()

	return outHolder.Out, nil
}

func (o *CommandExecOperation) renderTemplate(
	tmplName string,
	tmpl string,
	tmplData templateData,
	pf *files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (string, error) {
	parsedTmpl, tmplParseErr := template.New(tmplName).Parse(tmpl)
	if tmplParseErr != nil {
		if errorCh != nil {
			errorCh <- o.errorBuilder().Error(tmplParseErr)
		}

		if pf != nil {
			pf.SetFileProcessingError(
				NewCommandTemplateCanNotBeRenderedError(tmplParseErr),
			)
			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Failed(
					fmt.Sprintf("%s template can not be parsed. Template: %s", tmplName, tmpl),
					pf,
					tmplParseErr,
				)
			}
		}

		return "", tmplParseErr
	}

	var buf bytes.Buffer
	tmplExecErr := parsedTmpl.Execute(&buf, tmplData)
	if tmplExecErr != nil {
		if errorCh != nil {
			errorCh <- o.errorBuilder().Error(tmplExecErr)
		}

		if pf != nil {
			pf.SetFileProcessingError(
				NewCommandTemplateCanNotBeRenderedError(tmplExecErr),
			)
			if notificationCh != nil {
				notificationCh <- o.notificationBuilder().Failed(
					fmt.Sprintf("%s template can not be rendered. Template: %s", tmplName, tmpl),
					pf,
					tmplExecErr,
				)
			}
		}

		return "", tmplExecErr
	}

	return buf.String(), nil
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

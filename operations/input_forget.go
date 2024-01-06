package operations

import (
	"capyfile/files"
)

// InputForgetOperation forgets the input.
//
// Can be used when the current input should not be passed to the next operation.
// Combined with the file target policy, it can be used to forget certain input,
// such as the input with errors or without. Keep in mind that in such case, the
// cleanup policy won't be applied because capyfile stops tracking these files.
type InputForgetOperation struct {
	Name   string
	Params *InputForgetOperationParams
}

type InputForgetOperationParams struct {
}

func (o *InputForgetOperation) OperationName() string {
	return o.Name
}

func (o *InputForgetOperation) AllowConcurrency() bool {
	return false
}

func (o *InputForgetOperation) Handle(
	in []files.ProcessableFile,
	errorCh chan<- OperationError,
	notificationCh chan<- OperationNotification,
) (out []files.ProcessableFile, err error) {
	for i := range in {
		pf := &in[i]

		if notificationCh != nil {
			notificationCh <- o.notificationBuilder().Started("file is forgotten", pf)
		}
	}

	return out, err
}

func (o *InputForgetOperation) notificationBuilder() *OperationNotificationBuilder {
	return &OperationNotificationBuilder{
		OperationName: o.Name,
	}
}

func (o *InputForgetOperation) errorBuilder() *OperationErrorBuilder {
	return &OperationErrorBuilder{
		OperationName: o.Name,
	}
}

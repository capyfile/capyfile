package operations

import (
	"capyfile/files"
	"fmt"
)

const (
	StatusSkipped = iota
	StatusStarted
	StatusFinished
	StatusFailed
)

type OperationNotification struct {
	OperationName          string
	OperationStatus        int
	OperationStatusMessage string
	ProcessableFile        *files.ProcessableFile
	Error                  error
}

type OperationNotificationBuilder struct {
	OperationName string
}

func (nb *OperationNotificationBuilder) Skipped(message string, pf *files.ProcessableFile) OperationNotification {
	return OperationNotification{
		OperationName:          nb.OperationName,
		OperationStatus:        StatusSkipped,
		OperationStatusMessage: message,
		ProcessableFile:        pf,
	}
}

func (nb *OperationNotificationBuilder) Started(message string, pf *files.ProcessableFile) OperationNotification {
	return OperationNotification{
		OperationName:          nb.OperationName,
		OperationStatus:        StatusStarted,
		OperationStatusMessage: message,
		ProcessableFile:        pf,
	}
}

func (nb *OperationNotificationBuilder) Finished(message string, pf *files.ProcessableFile) OperationNotification {
	return OperationNotification{
		OperationName:          nb.OperationName,
		OperationStatus:        StatusFinished,
		OperationStatusMessage: message,
		ProcessableFile:        pf,
	}
}

func (nb *OperationNotificationBuilder) Failed(message string, pf *files.ProcessableFile, err error) OperationNotification {
	return OperationNotification{
		OperationName:          nb.OperationName,
		OperationStatus:        StatusFailed,
		OperationStatusMessage: message,
		ProcessableFile:        pf,
		Error:                  err,
	}
}

func NewOperationNotification(
	operationName string,
	status int,
	message string,
	pf *files.ProcessableFile,
	err error,
) OperationNotification {
	return OperationNotification{
		OperationName:          operationName,
		OperationStatus:        status,
		OperationStatusMessage: message,
		ProcessableFile:        pf,
		Error:                  err,
	}
}

func NewSkippedOperationNotification(
	operationName string,
	targetPolicy string,
	pf *files.ProcessableFile,
) OperationNotification {
	return NewOperationNotification(
		operationName,
		StatusSkipped,
		fmt.Sprintf("skipped due to \"%s\" target files policy", targetPolicy),
		pf,
		nil,
	)
}

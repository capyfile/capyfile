package files

type FileProcessingError interface {
	error
	Code() string
}

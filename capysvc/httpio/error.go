package httpio

type HTTPAwareType struct {
	httpStatusCode int

	code    string
	message string
	// Allows us building chained errors.
	errs []error
}

func NewHTTPAwareError(httpStatusCode int, code, message string, origErr error) *HTTPAwareType {
	return &HTTPAwareType{
		httpStatusCode: httpStatusCode,
		code:           code,
		message:        message,
		errs:           []error{origErr},
	}
}

func (e *HTTPAwareType) Error() string {
	return e.message
}

func (e *HTTPAwareType) HTTPStatusCode() int {
	return e.httpStatusCode
}

func (e *HTTPAwareType) Code() string {
	return e.code
}

func (e *HTTPAwareType) Message() string {
	return e.message
}

func (e *HTTPAwareType) OriginalError() error {
	return e.errs[0]
}

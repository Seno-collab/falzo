package share

type AppError struct {
	Code      string
	Message   string
	PublicErr error
	Internal  error
	Operation string
	Metadata  map[string]string
}

func NewAppError(code string, message string, publicErr error, internal error, operation string, metadata map[string]string) *AppError {
	return &AppError{
		Code:      code,
		Message:   message,
		PublicErr: publicErr,
		Internal:  internal,
		Operation: operation,
		Metadata:  metadata,
	}
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.PublicErr != nil {
		return e.PublicErr.Error()
	}
	return "application error"
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.PublicErr
}

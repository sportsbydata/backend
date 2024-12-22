package scouting

type NotFoundError struct {
	Message string
}

func (nfe *NotFoundError) Error() string {
	return nfe.Message
}

func NewNotFoundError(msg string) *NotFoundError {
	return &NotFoundError{
		Message: msg,
	}
}

type ValidationError struct {
	Message string
}

func (ve *ValidationError) Error() string {
	return ve.Message
}

func NewValidationError(msg string) *ValidationError {
	return &ValidationError{
		Message: msg,
	}
}

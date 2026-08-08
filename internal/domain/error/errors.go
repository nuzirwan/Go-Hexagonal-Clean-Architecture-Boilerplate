package domainerror

type Code string

const (
	ErrValidation   Code = "VALIDATION_ERROR"
	ErrNotFound     Code = "NOT_FOUND"
	ErrConflict     Code = "CONFLICT"
	ErrBusinessRule Code = "BUSINESS_RULE_VIOLATION"
	ErrUnauthorized Code = "UNAUTHORIZED"
)

type DomainError struct {
	Code    Code
	Message string
	Field   string
}

func (e *DomainError) Error() string { return string(e.Code) + ": " + e.Message }

func New(code Code, message string) *DomainError {
	return &DomainError{Code: code, Message: message}
}

func NewValidation(field, message string) *DomainError {
	return &DomainError{Code: ErrValidation, Message: message, Field: field}
}

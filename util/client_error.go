package util

// UnprocessableCode is an enum for UnprocessableError's Code field
type UnprocessableCode string

const (
	// CodeMissing means a resource does not exist
	CodeMissing UnprocessableCode = "missing"
	// CodeMissingField means a required field on a resource has not been set.
	CodeMissingField UnprocessableCode = "missing_field"
	// CodeInvalid means the formatting of a field is invalid
	CodeInvalid UnprocessableCode = "invalid"
	// CodeAlreadyExsits means another resource has the same value as this field.
	CodeAlreadyExsits UnprocessableCode = "already_exists"
)

// ClientError respond to 4xx http status.
type ClientError struct {
	Message string `json:"message"`
}

// InvalidReason respond to 422 status code
type InvalidReason struct {
	// Message is only used to pass data to the first argument of NewUnprocessable()
	Message string            `json:"message"`
	Field   string            `json:"field"`
	Code    UnprocessableCode `json:"code"`
}

// NewInvalidReason returns a new instance of InvalidReason.
func NewInvalidReason() *InvalidReason {
	return &InvalidReason{
		Message: "Validation failed",
	}
}

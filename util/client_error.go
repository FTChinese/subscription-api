package util

const (
	// CodeMissing means a resource does not exist
	CodeMissing = "missing"
	// CodeMissingField means a required field on a resource has not been set.
	CodeMissingField = "missing_field"
	// CodeInvalid means the formatting of a field is invalid
	CodeInvalid = "invalid"
	// CodeAlreadyExsits means another resource has the same value as this field.
	CodeAlreadyExsits = "already_exists"
	// CodeIncorrect means some field is not correct.
	CodeIncorrect = "incorrect"
)

// ClientError respond to 4xx http status.
type ClientError struct {
	Message string  `json:"message"`
	Reason  *Reason `json:"error,omitempty"`
}

// Reason tells why its unprocessable.
type Reason struct {
	message string
	Field   string `json:"field"`
	Code    string `json:"code"`
}

// NewReason creates a new instance of Reason
func NewReason() *Reason {
	return &Reason{message: "Validation failed"}
}

// SetMessage set the message to be carried away.
func (r *Reason) SetMessage(msg string) {
	r.message = msg
}

// GetMessage returns Reason's descriptive message.
func (r *Reason) GetMessage() string {
	return r.message
}

package aura

// ValidationError is returned when there is a validation error.
type ValidationError struct {
	err error
}

// Error stringifies the error.
func (e ValidationError) Error() string {
	if e.err == nil {
		return "validation error"
	}
	return e.err.Error()
}

// Unwrap returns the underlying error.
func (e ValidationError) Unwrap() error { return e.err }

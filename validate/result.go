package validate

import "errors"

// ValidationResult is a collection of FieldErrors with helper methods.
type ValidationResult struct {
	Errors FieldErrors
}

// NewResult returns a ValidationResult with specified errors.
//
// Use a ValidationResult to dynamically compose a set of validation errors.
func NewResult(errs ...error) *ValidationResult {
	result := &ValidationResult{}

	for _, err := range errs {
		if err != nil {
			result.AddError(err)
		}
	}

	return result
}

// IsValid returns false if one or more errors have been registered.
func (r *ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// Err returns FieldErrors.
func (r *ValidationResult) Err() error {
	if r.IsValid() {
		return nil
	}

	return r.Errors
}

func (r *ValidationResult) Error() string {
	if r.IsValid() {
		return ""
	}

	return r.Errors.Error()
}

// AddError adds an error to the ValidationResult.
func (r *ValidationResult) AddError(err error) {
	var fieldError FieldError

	var fieldErrors FieldErrors

	if errors.As(err, &fieldErrors) {
		for _, fieldErr := range fieldErrors {
			if err != nil {
				r.Errors = append(r.Errors, fieldErr)
			}
		}
	} else if errors.As(err, &fieldError) {
		r.Errors = append(r.Errors, fieldError)
	}
}

func (r *ValidationResult) AddErrors(err ...error) {
	for _, err := range err {
		r.AddError(err)
	}
}

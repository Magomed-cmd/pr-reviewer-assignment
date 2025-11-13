package validation

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrRequired = errors.New("value is required")
	ErrNil      = errors.New("value is nil")
)

type FieldError struct {
	Field  string
	Reason error
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Reason)
}

func (e FieldError) Unwrap() error {
	return e.Reason
}

func RequireString(field, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", FieldError{Field: field, Reason: ErrRequired}
	}

	return trimmed, nil
}

func RequireNotNil(field string, v any) error {
	if v == nil {
		return FieldError{Field: field, Reason: ErrNil}
	}

	return nil
}

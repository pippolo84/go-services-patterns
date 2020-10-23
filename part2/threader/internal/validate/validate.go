package validate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrJSONSyntax is a JSON syntax error
type ErrJSONSyntax struct {
	offset int64
}

// Error satisfies the error interface
func (e ErrJSONSyntax) Error() string {
	return fmt.Sprintf("Bad JSON at offset %d", e.offset)
}

// ErrInvalidValue is an invalid value error
type ErrInvalidValue struct {
	field  string
	offset int64
}

// Error satisfies the error interface
func (e ErrInvalidValue) Error() string {
	return fmt.Sprintf("Invalid value for the %q field at offset %d", e.field, e.offset)
}

// ErrUnknownField is an unknown field error
type ErrUnknownField struct {
	field string
}

// Error satisfies the error interface
func (e ErrUnknownField) Error() string {
	return fmt.Sprintf("JSON contains unknown field %s", e.field)
}

// ErrEmpty is an unknown field error
type ErrEmpty struct{}

// Error satisfies the error interface
func (e ErrEmpty) Error() string {
	return "JSON is empty"
}

// ErrSzExceeded is a body too large error
type ErrSzExceeded struct{}

// Error satisfies the error interface
func (e ErrSzExceeded) Error() string {
	return "JSON size too large"
}

// ErrGeneric is a generic unmarshalling error
type ErrGeneric struct{}

// Error satisfies the error interface
func (e ErrGeneric) Error() string {
	return "Generic error"
}

// ErrMultipleObjects is a multiple objects error
type ErrMultipleObjects struct{}

// Error satisfies the error interface
func (e ErrMultipleObjects) Error() string {
	return "Multiple objects"
}

// JSON validates the JSON inside a reader, unmarshalling it in a single object.
// It returns a non-nil error if the validation fails
func JSON(r io.Reader, obj interface{}) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&obj); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return ErrJSONSyntax{syntaxError.Offset}
		case errors.As(err, &unmarshalTypeError):
			return ErrInvalidValue{unmarshalTypeError.Field, unmarshalTypeError.Offset}
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			return ErrUnknownField{strings.TrimPrefix(err.Error(), "json: unknown field ")}
		case errors.Is(err, io.EOF):
			return ErrEmpty{}
		case strings.Contains(err.Error(), "too large"):
			return ErrSzExceeded{}
		default:
			return ErrGeneric{}
		}
	}

	// Call decode again to be sure that just a single JSON object is in
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return ErrMultipleObjects{}
	}

	return nil
}

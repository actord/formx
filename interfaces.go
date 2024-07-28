package formx

import (
	"context"
	"io"
)

// Component is the interface that all templates implement.
type Component interface {
	// Render the template.
	Render(ctx context.Context, w io.Writer) error
}

type FormValidation interface {
	ValidateStruct(s interface{})

	HasFormError() bool
	SetFormError(err string)
	GetFormError() string

	HasFieldError(field string) bool
	HasErrors() bool
	ErrorMessage(field string) string
	GetError(field string) FieldError
}

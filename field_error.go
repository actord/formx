package formx

// FieldError contains all functions to get error details
type FieldError interface {

	// ActualTag returns the validation tag that failed, even if an
	// alias the actual tag within the alias will be returned.
	// If an 'or' validation fails the entire or will be returned.
	//
	// eg. alias "iscolor": "hexcolor|rgb|rgba|hsl|hsla"
	// will return "hexcolor|rgb|rgba|hsl|hsla"
	ActualTag() string

	// Field returns the fields name with the tag name taking precedence over the
	// field's actual name.
	//
	// eq. JSON name "fname"
	// see StructField for comparison
	Field() string

	// Error returns the FieldError's message
	Error() string
}

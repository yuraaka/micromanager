package std

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func SetupGinValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Use JSON tag names in error messages
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return fld.Name
			}
			if name == "" {
				// Fallback to form tag
				name = strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
			}
			if name == "" {
				return fld.Name
			}
			return name
		})
	}
}

func FormatValidationErrors(errs validator.ValidationErrors) map[string]string {
	errors := make(map[string]string)

	for _, err := range errs {
		// err.Field() returns the struct field name
		// err.Tag() returns the validation tag that failed (e.g., "required", "min", "email")
		// err.Param() returns the parameter for the tag (e.g., "10" for "min=10")

		switch err.Tag() {
		case "required":
			errors[err.Field()] = "This field is required"
		case "email":
			errors[err.Field()] = "Invalid email format"
		case "min":
			errors[err.Field()] = "Must be at least " + err.Param() + " characters"
		case "max":
			errors[err.Field()] = "Must not exceed " + err.Param() + " characters"
		case "uuid":
			errors[err.Field()] = "Must be a valid UUID"
		default:
			errors[err.Field()] = "Failed validation: " + err.Tag()
		}
	}

	return errors
}

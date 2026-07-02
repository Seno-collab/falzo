package validatorx

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

var validate = newValidator()

func Validate(v any) []FieldError {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []FieldError{
			{
				Field:   "",
				Message: err.Error(),
			},
		}
	}

	result := make([]FieldError, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		result = append(result, FieldError{
			Field:   fieldErr.Field(),
			Message: messageFor(fieldErr),
		})
	}

	return result
}

func newValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return v
}

func messageFor(err validator.FieldError) string {
	field := err.Field()

	switch err.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "min":
		return field + " must be at least " + err.Param()
	case "max":
		return field + " must be at most " + err.Param()
	case "gte":
		return field + " must be greater than or equal to " + err.Param()
	case "lte":
		return field + " must be less than or equal to " + err.Param()
	case "oneof":
		return field + " must be one of: " + err.Param()
	default:
		return field + " is invalid"
	}
}

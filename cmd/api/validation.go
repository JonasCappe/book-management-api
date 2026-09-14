package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func validationFields(err error) map[string]string {
	fields := make(map[string]string)

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		fields["body"] = "request validation failed"
		return fields
	}

	for _, fieldError := range validationErrors {
		field := fieldError.Field()

		switch fieldError.Tag() {
		case "required":
			fields[field] = "is required"

		case "notblank":
			fields[field] = "must not be blank"

		case "max":
			fields[field] = fmt.Sprintf(
				"must not exceed %s characters",
				fieldError.Param(),
			)

		case "min":
			fields[field] = fmt.Sprintf(
				"must contain at least %s item(s)",
				fieldError.Param(),
			)

		case "unique":
			fields[field] = "must contain unique values"

		case "gt":
			fields[field] = fmt.Sprintf(
				"must be greater than %s",
				fieldError.Param(),
			)

		default:
			fields[field] = "is invalid"
		}
	}

	return fields
}

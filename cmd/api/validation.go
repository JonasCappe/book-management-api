package main

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func configureValidator(validate *validator.Validate) {
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	_ = validate.RegisterValidation("notblank", func(field validator.FieldLevel) bool {
		return strings.TrimSpace(field.Field().String()) != ""
	})
}

func (app *application) validatePayload(w http.ResponseWriter, r *http.Request, payload any) bool {
	err := app.validator.Struct(payload)
	if err == nil {
		return true
	}

	fields := make(map[string]string)
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		app.internalServerError(w, r, err)
		return false
	}
	for _, fieldError := range validationErrors {
		fields[fieldError.Field()] = validationMessage(fieldError)
	}
	app.validationError(w, r, fields)
	return false
}

func validationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "is required"
	case "notblank":
		return "must not be blank"
	case "max":
		return "must contain at most " + err.Param() + " characters"
	case "min":
		return "must contain at least " + err.Param() + " item"
	case "unique":
		return "must not contain duplicate values"
	case "gt":
		return "must contain only positive values"
	default:
		return "is invalid"
	}
}

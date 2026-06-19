package api

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

const (
	JSON = "json"
	FORM = "form"
)

type validatorError struct {
	FailedField string `json:"field"`
	Tag         string `json:"validation"`
	Value       string `json:"value"`
	Msg         string `json:"message"`
}

// supply struct with validator tags
func (api *Server) validate(s interface{}) []*validatorError {
	var errors []*validatorError
	err := api.valid.Struct(s)
	ref := reflect.ValueOf(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element validatorError
			// get JSON filed name by struct field name
			v, ok := ref.Type().FieldByName(err.Field())
			if ok {
				// get struct field name
				element.FailedField = err.Field()
				// get JSON filed name
				if v.Tag.Get(JSON) != "" {
					element.FailedField = v.Tag.Get(JSON)
				}
				// get FORM filed name
				if v.Tag.Get(FORM) != "" {
					element.FailedField = v.Tag.Get(FORM)
				}
			}
			element.Tag = err.Tag()
			element.Value = reflect.Indirect(ref).FieldByName(err.Field()).String()
			element.Msg = fmt.Sprintf("validation failed: invalid value '%s' in field '%s' failed on the '%s'", element.Value, element.FailedField, err.Tag())
			errors = append(errors, &element)
		}
	}
	return errors
}

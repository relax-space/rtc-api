package validator

import (
	"fmt"
	"nomni/utils/api"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"
)

type Validator struct {
	validator *validator.Validate
}

func (cv *Validator) Validate(i interface{}) error {
	err := cv.validator.Struct(i)
	if err == nil {
		return err
	}
	if errs, ok := err.(validator.ValidationErrors); ok {
		msg := make([]string, 0)
		for _, err := range errs {
			msg = append(msg, fmt.Sprintf("%v condition: %v ,value: %v", err.Field(), err.ActualTag(), err.Value()))
		}
		return api.InvalidFieldError(strings.Join(msg, ","))
	}
	return err
}
func New() *Validator {
	return &Validator{validator: validator.New()}
}

package validator

import (
	"errors"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validator *validator.Validate
}

func New() *Validator {
	return &Validator{validator: validator.New()}
}

func (v *Validator) Validate(request any) (err error) {
	err = v.validator.Struct(request)
	if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			return validateErrs
		}
	}
	return err
}

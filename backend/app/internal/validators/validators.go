package validators

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

func ValidateCustomEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	match, _ := regexp.MatchString(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`, email)
	return match
}

func ValidateCustomPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	match, _ := regexp.MatchString(`^\+7[0-9]{10}$`, phone)
	return match
}

func RegisterCustomValidations(v *validator.Validate) error {
	if err := v.RegisterValidation("customemail", ValidateCustomEmail); err != nil {
		return err
	}
	if err := v.RegisterValidation("customphone", ValidateCustomPhone); err != nil {
		return err
	}
	return nil
}

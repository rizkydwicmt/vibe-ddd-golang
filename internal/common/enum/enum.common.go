package enum

import (
	"github.com/go-playground/validator/v10"
)

type Enum interface {
	ToString() string
	IsValid() bool
}

func ValidateEnum(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(Enum)
	if !ok {
		return false
	}
	return value.IsValid()
}

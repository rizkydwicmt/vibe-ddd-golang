package types

import (
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

type StringToBool string

func (s StringToBool) ToBool() bool {
	value, _ := strconv.ParseBool(string(s))
	return value
}

func ValidateStringToBool(fl validator.FieldLevel) bool {
	value := fl.Field().Interface().(StringToBool)
	_, err := strconv.ParseBool(strings.ToLower(string(value)))
	return err == nil
}
